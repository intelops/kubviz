CREATE TABLE IF NOT EXISTS jfrogcontainerpush (
	Domain         String,
	EventType      String,
	RegistryURL    String,
	RepositoryName String,
	SHAID          String,
	Size           Int32,
	ImageName      String,
	Tag            String,
	Event          String,
    EventTime      DateTime('UTC'),
	ExpiryDate DateTime DEFAULT now() + INTERVAL {{.TTLValue}} {{.TTLUnit}},
	ExportedAt DateTime DEFAULT NULL
) ENGINE = MergeTree() 
ORDER BY ExpiryDate 
TTL ExpiryDate;

