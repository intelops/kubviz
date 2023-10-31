CREATE TABLE IF NOT EXISTS DeprecatedAPIs (
	ClusterName     String,
	ObjectName      String,
	Description     String,
	Kind            String,
	Deprecated      UInt8,
	Scope           String,
	EventTime       DateTime('UTC')
) engine=File(TabSeparated);
	
