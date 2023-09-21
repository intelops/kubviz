CREATE TABLE IF NOT EXISTS dockerhubbuild (
	PushedBy      String,
	ImageTag      String,
	RepositoryName String,
	DateCreated   String,
	Owner         String,
	Event         String,
    EventTime     DateTime('UTC')
) engine=File(TabSeparated);
