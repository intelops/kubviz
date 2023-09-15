CREATE TABLE IF NOT EXISTS getall_resources (
	ClusterName String,
	Namespace   String,
	Kind        String,
	Resource    String,
	Age         String,
    EventTime   DateTime('UTC')
) engine=File(TabSeparated);
