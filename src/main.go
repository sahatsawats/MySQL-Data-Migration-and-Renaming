package main

import (
	"MDMR/src/conf"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"os/exec"
	concurrentlog "github.com/sahatsawats/concurrent-log"
	concurrentqueue "github.com/sahatsawats/concurrent-queue"
	"gopkg.in/yaml.v2"
)

// TODO: Make the directory if not exists
func makeDirectory(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.Mkdir(path, 0755)

		if err != nil {
			return err
		}
	}

	return nil
}

// TODO: Reading configuration file from ./conf/config.yaml based on executable path
func readingConfigurationFile() *mdmr_config.Configurations {
	// get the current execute directory
	baseDir, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}

	// Join path to config file
	configFile := filepath.Join(filepath.Dir(baseDir), "conf", "config.yaml")
	// Read file in bytes for mapping yaml to structure with yaml package
	readConf, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Failred to read configuration file: %v", err)
	}

	// Map variable to configuration function
	var conf mdmr_config.Configurations
	// Map yaml file to config structure
	err = yaml.Unmarshal(readConf, &conf)
	if err != nil {
		log.Fatalf("Failed to unmarshal config file: %v", err)
	}

	return &conf
}

func dumpSchemaByHost(id int, wg *sync.WaitGroup, host string, conf *mdmr_config.Configurations, logHandler *concurrentlog.Logger) {
	// set postpone to issued the done signal to wait group
	defer wg.Done()

	// Split the ipadress and port out of string
	hostCredentials := strings.Split(host, ":")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/", conf.Database.SourceDBUser, conf.Database.SourceDBPassword, hostCredentials[0], hostCredentials[1])

	// Preparing the MySQL Connection.
	logHandler.Log("INFO", fmt.Sprintf("[thread:%d] Start open connection with %s", id, dsn))

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		logHandler.Log("ERROR", fmt.Sprintf("Failed to create database connection pool: %v", err))
		os.Exit(1)
	}

	err = db.Ping()
	if err != nil {
		logHandler.Log("ERROR", fmt.Sprintf("Failed to connect the database: %v", err))
	}

	// Collect time consuming
	enqueueStartTime := time.Now()
	// SQL statement to query list of databases execept system databases.
	queryStatement := fmt.Sprintf("SELECT schema_name FROM information_schema.schemata WHERE schema_name LIKE '%s%%';", conf.MDMR.SourcePrefix)

	// Querying the list of databases
	rows, err := db.Query(queryStatement)
	if err != nil {
		logHandler.Log("ERROR", fmt.Sprintf("cannot query from database: %v", err))
	}

	// Declare the concurrent queue for hold database name.
	queue := concurrentqueue.New[string]()
	// Loop over the results
	var err_enq int
	for rows.Next() {
		var databaseName string // Declare variable for holding the results
		err = rows.Scan(&databaseName) // Map the value of results to variable
		if err != nil {
			logHandler.Log("ERROR", fmt.Sprintf("Failed to read result: %v", err))
			// Collect the metrics
			err_enq += 1
		}

		// Enqueue the data
		queue.Enqueue(databaseName)

	}
	enqueueElaspedTime := time.Since(enqueueStartTime)
	logHandler.Log("INFO", fmt.Sprintf("[thread:%d] Complete enqueue database name with time usage: %v and error reported: %d", id, enqueueElaspedTime, err_enq))

	// Close the database connection
	rows.Close()

	// Start the dumpThreads
	var wg_dump sync.WaitGroup
	dumpThreads := conf.MDMR.DumpThreads

	for i := 0; i < dumpThreads; i++ {
		wg_dump.Add(1)
		go func() {
			defer wg_dump.Done()
			for {
				// Break condition
				if queue.IsEmpty() {
					return
				}
				// Dequeue database name from queue
				databaseName := queue.Dequeue()

				// Mapping the staging directory and staging file name
				stagingDir := conf.MDMR.StagingDirectory
				stagingFileName := filepath.Join(stagingDir, fmt.Sprintf("%s-staging", databaseName))

				// mysqlsh -h <IP> -P <PORT> -u <user> -p'<pwd>' -e "util.dumpSchemas(['<database_name>'], {thteads: 4})"
				cmd := exec.Command(
					"mysqlsh", "-h", hostCredentials[0], "-P", hostCredentials[1], 
					"-u", conf.Database.SourceDBUser, 
					fmt.Sprintf("-p'%s'", conf.Database.SourceDBPassword), 
					"-e", fmt.Sprintf("util.dumpSchemas(['%s'], '%s', {threads: 4})", databaseName, stagingFileName))

				err := cmd.Run()
				if err != nil {
					logHandler.Log("ERROR", fmt.Sprintf("Failed to execute dumpSchemas command from host: %s with err_statement: %v", hostCredentials[0], err))
				}
			}
		}()
	}
}






func main() {
	fmt.Println("Start reading configuration file...")
	mdmr_config := readingConfigurationFile()
	fmt.Println("Complete reading configuration file.")
	fmt.Println("Starting logging thread...")
	logPath := filepath.Join(mdmr_config.Logger.LogDirectory, mdmr_config.Logger.LogFileName)
	// Create concurrent logger
	logHandler, err := concurrentlog.NewLogger(logPath, 50)
	if err != nil {
		log.Fatalf("Failed to initialize log handler: %v", err)
	}

	fmt.Println("Complete create logging thread. Starting logging...")

	// Create staging directory for holding the dump file
	err = makeDirectory(mdmr_config.MDMR.StagingDirectory)
	if err != nil {
		logHandler.Log("ERROR", fmt.Sprintf("Failed to create directory: %v", err))
		os.Exit(1)
	}
	// Spliting the source hosts to list of host
	sourceHostList := strings.Split(mdmr_config.Server.SourceAddress, ",")
	logHandler.Log("INFO", fmt.Sprintf("Source Host: %s, total: %d", sourceHostList, len(sourceHostList)))

	// Create wait group for dump operation
	var wg sync.WaitGroup
	for i := 1; i <= len(sourceHostList); i++ {
		wg.Add(1)
		go dumpSchemaByHost(i, &wg, sourceHostList[i-1], mdmr_config, logHandler)
	}

	wg.Wait()




}