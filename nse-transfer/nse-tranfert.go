package main

import (
	"bufio"
	"fmt"
	"gnuzip"
	"io"
	"log"
	"os"
	"path"
	"regexp"
	"sftpclient"
	"strconv"
	"time"
)

//const SFTP_ADDR = "10.1.151.60"
//const SFTP_PORT = "22"
const SFTP_ADDR = "10.5.8.19"
const SFTP_PORT = 5222
const SFTP_USER = "sftpbot"
const SFTP_PATH = "sftpbot"
const SFTP_RSAKEY = "/home/geoffrey.mathy/id_sftpbot.rsa"

const CFGPATH = "/etc/nse-collect.conf"
const EVDIR = "/var/spool/nagios/events-bt"
const EVPATH = "/var/spool/nagios/events-bt/events.out"
const LOG = "/var/log/nagios/events-bt/transfert.log"

/*
 * Parse config file
 */
func parseConfigFile(cfgpath string) (map[string]string, error) {
	cfg := make(map[string]string)

	cf, err := os.Open(CFGPATH)
	if err != nil {
		return nil, fmt.Errorf("unable to open config file: %v", err)
	}
	defer cf.Close()

	scan := bufio.NewScanner(cf)
	for scan.Scan() {
		re := regexp.MustCompile(`^\s*([a-zA-Z0-9-_]+)\s*=\s*([a-zA-Z0-9-_./]+)\s*$`)
		substr := re.FindStringSubmatch(scan.Text())
		//- processing config line
		if len(substr) != 3 {
			return nil, fmt.Errorf("unable to correctly parse this configuration line: %s. %v", scan.Text(), substr)
		}
		varname := substr[1]
		varvalue := substr[2]
		cfg[varname] = varvalue
	}
	if err := scan.Err(); err != nil {
		fmt.Println("error")
	}
	return cfg, nil
}

func main() {
	//- Parse configuration file
	c, err := parseConfigFile(CFGPATH)
	if err != nil {
		log.Fatal(err)
	}

	//- Init logger
	logfile, err := os.OpenFile(LOG, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640)
	if err != nil {
		fmt.Errorf("unable to open log file: ", err)
	}
	defer logfile.Close()

	mw := io.MultiWriter(os.Stdout, logfile)

	logger := log.New(mw, "", log.LstdFlags)
	logger.Println("--- CKECK FILE FOR TRANSFERT -----------------------")
	logger.SetFlags(0)

	//- Check is the EVPATH file exists
	if _, err := os.Stat(EVPATH); os.IsNotExist(err) {
		logger.Println("no file to process")
		os.Exit(0)
	}

	// STEP 1 - COMPRESS
	zip, _ := os.Hostname()
	zip += "_"
	zip += strconv.FormatInt(time.Now().Unix(), 10)
	zip += ".gz"
	zip = path.Join(EVDIR, zip)

	logger.Printf("compressing %s to %s", EVPATH, zip)
	err = gnuzip.Compress(zip, EVPATH)
	if err != nil {
		log.Fatal(err)
	}

	// STEP 2 - OPEN SFTP CONNECTION
	logger.Printf("pushing gzip file to SFTP %s:%s", c["sftpserver"], c["sftpport"])
	sftp, err := SftpClient.New(c["sftpserver"],
		c["sftpport"],
		c["sftpuser"],
		c["sftppath"],
		c["rsakey"],
		true)
	if err != nil {
		log.Fatal(err)
	}

	sftp.PutFiles([]string{zip})
	if err != nil {
		log.Fatal(err)
	}
	sftp.Close()

	//- Remove event file
	logger.Println("removing files")
	if err = os.Remove(EVPATH); err != nil {
		logger.Fatalf("unable to delete event file: %v", err)
	}
	if err = os.Remove(zip); err != nil {
		logger.Fatalf("unable to delete event file: %v", err)
	}
}
