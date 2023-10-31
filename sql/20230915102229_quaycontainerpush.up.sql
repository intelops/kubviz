CREATE TABLE IF NOT EXISTS quaycontainerpush (
	name          String,
	repository    String,
	nameSpace     String,
	dockerURL     String,
	homePage      String,
	tag           String,
	Event         String,
    EventTime DateTime('UTC')
) engine=File(TabSeparated);
