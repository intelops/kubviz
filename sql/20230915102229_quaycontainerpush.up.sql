CREATE TABLE IF NOT EXISTS quaycontainerpush (
	name          String,
	repository    String,
	nameSpace     String,
	dockerURL     String,
	homePage      String,
	tag           String,
	Event         String,
    EventTime DateTime('UTC'),
	ExpiryDate DateTime DEFAULT now() + INTERVAL 1 MONTH
) ENGINE = MergeTree() 
ORDER BY ExpiryDate 
TTL ExpiryDate;
