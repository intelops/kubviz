CREATE TABLE IF NOT EXISTS azurecontainerpush (
	RegistryURL    String,
	RepositoryName String,
	Tag            String,
	ImageName      String,
	Event          String,
	Size           Int32,
	SHAID          String,
    EventTime DateTime('UTC'),
	ExpiryDate DateTime DEFAULT now() + INTERVAL 1 MONTH
) ENGINE = MergeTree() 
ORDER BY ExpiryDate 
TTL ExpiryDate;