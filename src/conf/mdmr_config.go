package mdmr_config

	type Configurations struct {
		Server		ServerConfigurations
		Database	DatabaseConfigurations
		Logger		LoggerConfigurations
		MDMR	SoftwareConfigurations
	}

	type ServerConfigurations struct {
		// SourceAddress: "ipaddress1:port1, ipaddress2:port2,..."
		SourceAddress string
		// DestinationAddress: "ipaddress:port"
		DestinationAddress string
	}

	type DatabaseConfigurations struct {
		SourceDBUser string
		SourceDBPassword string
		DestinationDBUser string
		DestinationDBPassword string
	}

	type LoggerConfigurations struct {
		LogDirectory string
		LogFileName string
	}

	type SoftwareConfigurations struct {
		SourcePrefix string
		DestinationPrefix string
		Threads int
	}