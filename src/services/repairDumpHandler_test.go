package services_test

import (
	"MDMR/src/models"
	"MDMR/src/services"
	"path/filepath"
	"testing"
	"os"
	"time"
)



func TestRepairHandler(t *testing.T) {
	// Temporary directory for testing
	//tmpDir := t.TempDir()
	tmpDir := "/Users/sahatsawat/Projects/MDMR/test/services/repairDumpHandler"
	logFilePath := filepath.Join(tmpDir, "repair.log")

	// Create a RepairHandler
	repairHandler := services.NewRepairHandler(logFilePath, tmpDir, 10)
	defer repairHandler.Close()

	// Mock MySQL credentails
	mockCredentials := models.MySQLCredentials{
		Host:	"213.35.99.8",
		Port:	"3306",
		User: 	"root",
		Password: "M@Tc#2024",
	}

	// Create a mock RepairTask
	mockTask := models.RepairTask{
		DatabaseName:    "admin_mtls_biz_002",
		MySQLCredentials: mockCredentials,
	}

	// Channel to capture the output
	done := make(chan bool)

	go func() {
		repairHandler.Repair("admin_mtls_biz_002", mockTask.MySQLCredentials)
		repairHandler.Repair("admin_mtls_biz_003", mockTask.MySQLCredentials)
		repairHandler.Repair("admin_mtls_biz_004", mockTask.MySQLCredentials)
		repairHandler.Repair("admin_mtls_biz_005", mockTask.MySQLCredentials)
		repairHandler.Repair("admin_mtls_biz_006", mockTask.MySQLCredentials)
		done <- true
	}()
	
	select {
	case <-done:
		time.Sleep(5000 * time.Millisecond)

		_, err := os.Stat(logFilePath)
		// Check if the log file was created
		if os.IsNotExist(err) {
			t.Fatalf("Expected log file at %s but it does not exist", logFilePath)
		}


		expectedDirectory := filepath.Join(tmpDir,"admin_mtls_biz_002-staging-repair")

		if !isDirectoryExists(expectedDirectory){
			t.Errorf("Error Does not found the directory: %s", expectedDirectory)
		}
		
	case <-time.After(2 * time.Second):
		t.Fatalf("Test timed out waiting for repair to be processed")
	}
}


// Helper function to check if a string is contained in another string
func isDirectoryExists(filePath string) bool {
	_, err := os.Stat(filePath)

	return !os.IsNotExist(err)
}