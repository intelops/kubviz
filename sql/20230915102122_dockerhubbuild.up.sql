CREATE TABLE IF NOT EXISTS dockerhubbuild (
	PushedBy      String,
	ImageTag      String,
	RepositoryName String,
	DateCreated   String,
	Owner         String,
	Event         String,
    EventTime     DateTime('UTC'),
	ExpiryDate DateTime DEFAULT now() + INTERVAL 6 MONTH
) ENGINE = MergeTree() 
ORDER BY ExpiryDate 
TTL ExpiryDate;
