package services

import (
	"MDMR/src/models"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/natefinch/lumberjack"
)

type RepairHandler struct {
	stdLog           *log.Logger
	repairChan       chan models.RepairTask
	done             chan struct{}
	repairStagingDir string
	wg				sync.WaitGroup
}

// TODO: Construct the RepairHandler
func NewRepairHandler(logFilePath string, stagingDir string, bufferSize int) *RepairHandler {
	// Create a logger instance
	repairLogger := &lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    10,
		MaxBackups: 0,
		MaxAge:     28,
	}

	stdLogger := log.New(repairLogger, "RepairHandles: ", log.LstdFlags)

	// Create RepairHandler object
	repairObject := &RepairHandler{
		stdLog:              stdLogger,
		repairChan:       make(chan models.RepairTask, bufferSize),
		done:             make(chan struct{}),
		repairStagingDir: stagingDir,
	}

	// Initialize the RepairObject
	go repairObject.run()

	// Return the object
	return repairObject
}

// TODO: Start run the process.
func (r *RepairHandler) run() {
	// Infinity loop until received signal
	for {
		select {
		case repairTask := <-r.repairChan:
			r.stdLog.Printf("Starting repair task from database: %s", repairTask.DatabaseName)
			// <database_name>-staging-repair
			stagingFileName := fmt.Sprintf("%s-staging-repair", repairTask.DatabaseName)
			// Map the staging file name with path
			stagingPath := filepath.Join(r.repairStagingDir, stagingFileName)
			cmd := exec.Command(
				"mysqlsh", "-h", repairTask.MySQLCredentials.Host, "-P", repairTask.MySQLCredentials.Port,
				"-u", repairTask.MySQLCredentials.User,
				fmt.Sprintf("-p%s", repairTask.MySQLCredentials.Password),
				"--js",
				"-e", fmt.Sprintf("util.dumpSchemas(['%s'], '%s', {threads: 4})", repairTask.DatabaseName, stagingPath))

			err := cmd.Run()
			if err != nil {
				r.stdLog.Printf("Failed to retry database name: %s with error: %v \n", repairTask.DatabaseName, err)
			} else {
				r.stdLog.Printf("Completed retry database name: %s from %s \n", repairTask.DatabaseName, repairTask.MySQLCredentials.Host)
			}
			r.wg.Done()
		case <-r.done:
			return
		}
	}
}

// TODO: Act as API for send the repairing request to object
func (r *RepairHandler) Repair(databaseName string, credentials models.MySQLCredentials) {
	r.wg.Add(1)
	r.stdLog.Printf("Receiving repair task: %s from %s \n", databaseName, credentials.Host)
	// Create a repairTask with databaseName and Credentails
	repairTask := models.RepairTask{
		DatabaseName:     databaseName,
		MySQLCredentials: credentials,
	}
	// Send the repairTask to channel
	r.repairChan <- repairTask
}

// TODO: Close the buffer channel
func (r *RepairHandler) Close() {
	r.wg.Wait()
	close(r.done)
	close(r.repairChan)
	
}
