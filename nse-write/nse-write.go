package main

import (
	"encoding/json"
	"github.com/gofrs/flock"
	"io"
	"log"
	"os"
)

const OUT = "/var/spool/nagios/events/events.out"
const LOG = "/var/log/nagios/events/writer.log"

type JsonEvent struct {
	Perimeter string
	Host      string
	Service   string
	Output    string
	State     string
	StateType string
	Attempt   string
	Datetime  string
}

type NagiosWriter struct {
	logFile *os.File
	outFile *os.File
	outPath string
	logger  *log.Logger
	json    []byte
}

func (nw *NagiosWriter) Init(lf, of string) error {
	var err error

	//- Init logger
	(*nw).outPath = of
	(*nw).logFile, err = os.OpenFile(lf, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640)
	if err != nil {
		return err
	}
	mw := io.MultiWriter(os.Stdout, (*nw).logFile)
	(*nw).logger = log.New(mw, "", log.LstdFlags)

	return nil
}

func (nw *NagiosWriter) parseDataToJson(je *JsonEvent) (bool, error) {
	var err error

	//- Only handle SOFT states
	if je.StateType != "SOFT" && je.StateType != "soft" {
		return false, nil
	}

	//- Json Marshal
	(*nw).json, err = json.Marshal(*je)
	if err != nil {
		return true, err
	}
	return true, nil
}

func (nw *NagiosWriter) Println(s string) {
	(*nw.logger).Println(s)
}

func (nw *NagiosWriter) LoggerPrefix(b bool) {
	if b {
		(*nw.logger).SetFlags(log.LstdFlags)
	} else {
		(*nw.logger).SetFlags(0)
	}
}

func (nw *NagiosWriter) Printf(s string, a ...interface{}) {
	(*nw.logger).Printf(s, a...)
}

func (nw *NagiosWriter) Fatalf(s string, a ...interface{}) {
	(*nw.logger).Fatalf(s, a...)
}

func (nw *NagiosWriter) WriteEvent() error {
	var err error

	//- Open output file
	(*nw).outFile, err = os.OpenFile((*nw).outPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640)
	if err != nil {
		return err
	}
	//- Locking event file
	lock := flock.New((*nw).outPath)
	for {
		locked, err := lock.TryLock()
		if locked {
			break
		}
		if err != nil {
			return err
		}
	}

	//- Write into event file
	jsonLine := append((*nw).json, "\n"...)
	if _, err = (*nw).outFile.Write(jsonLine); err != nil {
		return err
	}

	//- Remove lock
	if err := lock.Unlock(); err != nil {
		return err
	}
	return nil
}

func (nw *NagiosWriter) Close() {
	(*nw).logFile.Close()
	(*nw).outFile.Close()
}

func main() {
	if len(os.Args) < 9 {
		log.Fatal("invalid number of args")
	}

	/**
	 * Initiate logger
	 */
	var nwriter = new(NagiosWriter)
	if err := nwriter.Init(LOG, OUT); err != nil {
		log.Fatalf("unable to init NagiosWriter: %v", err)
	}

	/**
	 * Create JSON string from args
	 */
	notif := JsonEvent{
		os.Args[1],
		os.Args[2],
		os.Args[3],
		os.Args[4],
		os.Args[5],
		os.Args[6],
		os.Args[7],
		os.Args[8],
	}

	//- First logs args
	nwriter.Println("--- NEW EVENT ------------------------------------------------------------------------------")
	nwriter.LoggerPrefix(false)
	nwriter.Printf("Perimeter:%s, Host:%s, Service:%s, Output:%s, State:%s, StateType:%s, Attempt:%s, Datetime:%s ",
		os.Args[1], os.Args[2], os.Args[3], os.Args[4], os.Args[5], os.Args[6], os.Args[7], os.Args[8])

	valid, err := nwriter.parseDataToJson(&notif)
	if !valid {
		nwriter.Println("NON-SOFT state are ignored")
		os.Exit(0)
	} else if err != nil {
		nwriter.Fatalf("unable to parse json data: %v", err)
	}

	/**
	 * Append string to event file
	 */
	if err = nwriter.WriteEvent(); err != nil {
		nwriter.Fatalf("unable to write json into event file: %v", err)
	}

	nwriter.Println("event processed successfully")
	nwriter.Close()
}
