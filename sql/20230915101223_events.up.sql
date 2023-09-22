CREATE TABLE IF NOT EXISTS events (
	ClusterName String,
	Id          String,
	EventTime   DateTime('UTC'),
	OpType      String,
	Name        String,
	Namespace   String,
	Kind        String,
	Message     String,
	Reason      String,
	Host        String,
	Event       String,
	FirstTime   String,
	LastTime    String
) engine=File(TabSeparated);
