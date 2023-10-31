CREATE TABLE IF NOT EXISTS kubescore (
	id              UUID,
	namespace       String,
	cluster_name    String,
	recommendations String,
    EventTime       DateTime('UTC')
) engine=File(TabSeparated);