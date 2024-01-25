CREATE TABLE IF NOT EXISTS dockerhubbuild (
	PushedBy      String,
	ImageTag      String,
	RepositoryName String,
	DateCreated   String,
	Owner         String,
	Event         String,
    EventTime     DateTime('UTC'),
	ExpiryDate DateTime DEFAULT now() + INTERVAL {{.TTLValue}} {{.TTLUnit}},
    ExportedAt  DateTime DEFAULT NULL
) ENGINE = MergeTree() 
ORDER BY ExpiryDate 
TTL ExpiryDate;
