# nse-collect (Nagios Soft Event)
Collects all Nagios SOFT event and publishes then on a SFTP server

This project builds on other project that are here to avoid compatibily issues. But they can be installed this way : 
* go get "github.com/pkg/sftp"
* go get "golang.org/x/crypto/ssh"
* go get -u "github.com/gofrs/flock"

NSE-Collect provides two components : write & transfer

## nse-write
Must be called by Nagios event handler. It retrieves and format its parameters and finally writes them into a JSON file.
* Spool file : /var/spool/nagios/events/events.out
* Log file : /var/log/nagios/events/writer.log

## nse-transfer
Must be called periodically by a cron task. It compresses JSON data to Gzip file and transfers them on an SFTP server. If the transfer failed, it will be retried the next time. 
* Spool file to compress : /var/spool/nagios/events/events.out (same than nse-writer)
* Log file : /var/log/nagios/events/transfert.log

## Installation
You have to : 
1. Compile and put nse-writer and nse-transfer into /usr/local/bin/nse/ (to compile, *go build nse-write* and *go build nse-transfer*)
2. Create directory /var/spool/nagios/events (owner nagios:nagios / mode 750)
3. Create directory /var/log/nagios/events (owner nagios:nagios / mode 750)
4. Add this configuration in /etc/nagios/nagios.conf :
```
define command{
command_name softevent-host
command_line  /usr/local/bin/nse/nse-write '$HOSTGROUPALIAS$' '$HOSTALIAS$' 'none' '$HOSTOUTPUT$' '$HOSTSTATE$' '$HOSTSTATETYPE$' '$HOSTATTEMPT$' '$SHORTDATETIME$'
}
define command{
command_name softevent-service
command_line  /usr/local/bin/nse/nse-write '$HOSTGROUPALIAS$' '$HOSTALIAS$' '$SERVICEDESC$' '$SERVICEOUTPUT$' '$SERVICESTATE$' '$SERVICESTATETYPE$' '$SERVICEATTEMPT$' '$SHORTDATETIME$'
}
```

## Configuration
Copy the file etc/nse-collect.conf in this repository in the /etc directory of your server. Edit it to adjust values. 
