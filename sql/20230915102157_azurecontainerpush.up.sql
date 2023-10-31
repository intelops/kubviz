CREATE TABLE IF NOT EXISTS azurecontainerpush (
	RegistryURL    String,
	RepositoryName String,
	Tag            String,
	ImageName      String,
	Event          String,
	Size           Int32,
	SHAID          String,
    EventTime DateTime('UTC')
) engine=File(TabSeparated);
