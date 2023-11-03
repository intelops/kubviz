CREATE TABLE IF NOT EXISTS dockerhubbuild (
	PushedBy      String,
	ImageTag      String,
	RepositoryName String,
	DateCreated   String,
	Owner         String,
	Event         String,
    EventTime     DateTime('UTC'),
	ExpiryDate DateTime DEFAULT now() + INTERVAL 1 MONTH
) ENGINE = MergeTree() 
ORDER BY ExpiryDate 
TTL ExpiryDate;
