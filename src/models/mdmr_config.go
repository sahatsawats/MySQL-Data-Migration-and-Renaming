package models

	type Configurations struct {
		Server		ServerConfigurations
		Database	DatabaseConfigurations
		Logger		LoggerConfigurations
		MDMR	SoftwareConfigurations
	}

	type ServerConfigurations struct {
		// SourceAddress: "ipaddress1:port1, ipaddress2:port2,..."
		SourceAddress string
	}

	type DatabaseConfigurations struct {
		SourceDBUser string
		SourceDBPassword string
	}

	type LoggerConfigurations struct {
		LogDirectory string
		LogFileName string
	}

	type SoftwareConfigurations struct {
		SourcePrefix string
		DumpThreads int
		// stagingDirectory: /data/path
		StagingDirectory string
		// RepairStaging for repairing process
		RepairStagingDirectory string
	}