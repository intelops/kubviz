-- kubvizTable
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

-- rakeesTable
CREATE TABLE IF NOT EXISTS rakkess (
	ClusterName String,
	Name        String,
	Create      String,
	Delete      String,
	List        String,
	Update      String
) engine=File(TabSeparated);

-- kubePugDepricatedTable
CREATE TABLE IF NOT EXISTS DeprecatedAPIs (
	ClusterName String,
	ObjectName  String,
	Description String,
	Kind        String,
	Deprecated  UInt8,
	Scope       String
) engine=File(TabSeparated);

-- kubepugDeletedTable
CREATE TABLE IF NOT EXISTS DeletedAPIs (
	ClusterName String,
	ObjectName  String,
	Group       String,
	Kind        String,
	Version     String,
	Name        String,
	Deleted     UInt8,
	Scope       String
) engine=File(TabSeparated);

-- jfrogContainerPushEventTable
CREATE TABLE IF NOT EXISTS jfrogcontainerpush (
	Domain         String,
	EventType      String,
	RegistryURL    String,
	RepositoryName String,
	SHAID          String,
	Size           Int32,
	ImageName      String,
	Tag            String,
	Event          String
) engine=File(TabSeparated);

-- ketallTable
CREATE TABLE IF NOT EXISTS getall_resources (
	ClusterName String,
	Namespace   String,
	Kind        String,
	Resource    String,
	Age         String
) engine=File(TabSeparated);

-- outdateTable
CREATE TABLE IF NOT EXISTS outdated_images (
	ClusterName     String,
	Namespace       String,
	Pod             String,
	CurrentImage    String,
	CurrentTag      String,
	LatestVersion   String,
	VersionsBehind  Int64
) engine=File(TabSeparated);

-- kubescoreTable
CREATE TABLE IF NOT EXISTS kubescore (
	id             UUID,
	namespace      String,
	cluster_name   String,
	recommendations String
) engine=File(TabSeparated);

-- trivyTableVul
CREATE TABLE IF NOT EXISTS trivy_vul (
	id                    UUID,
	cluster_name          String,
	namespace             String,
	kind                  String,
	name                  String,
	vul_id                String,
	vul_vendor_ids        String,
	vul_pkg_id            String,
	vul_pkg_name          String,
	vul_pkg_path          String,
	vul_installed_version String,
	vul_fixed_version     String,
	vul_title             String,
	vul_severity          String,
	vul_published_date    DateTime('UTC'),
	vul_last_modified_date DateTime('UTC')
) engine=File(TabSeparated);

-- trivyTableMisconfig
CREATE TABLE IF NOT EXISTS trivy_misconfig (
	id                 UUID,
	cluster_name       String,
	namespace          String,
	kind               String,
	name               String,
	misconfig_id       String,
	misconfig_avdid    String,
	misconfig_type     String,
	misconfig_title    String,
	misconfig_desc     String,
	misconfig_msg      String,
	misconfig_query    String,
	misconfig_resolution String,
	misconfig_severity String,
	misconfig_status   String
) engine=File(TabSeparated);

-- trivyTableImage
CREATE TABLE IF NOT EXISTS trivyimage (
	id                  UUID,
	cluster_name        String,
	artifact_name       String,
	vul_id              String,
	vul_pkg_id          String,
	vul_pkg_name        String,
	vul_installed_version String,
	vul_fixed_version   String,
	vul_title           String,
	vul_severity        String,
	vul_published_date  DateTime('UTC'),
	vul_last_modified_date DateTime('UTC')
) engine=File(TabSeparated);

-- dockerHubBuildTable
CREATE TABLE IF NOT EXISTS dockerhubbuild (
	PushedBy      String,
	ImageTag      String,
	RepositoryName String,
	DateCreated   String,
	Owner         String,
	Event         String
) engine=File(TabSeparated);

-- azureContainerPushEventTable
CREATE TABLE IF NOT EXISTS azurecontainerpush (
	RegistryURL    String,
	RepositoryName String,
	Tag            String,
	ImageName      String,
	Event          String,
	Timestamp      String,
	Size           Int32,
	SHAID          String
) engine=File(TabSeparated);

-- quayContainerPushEventTable
CREATE TABLE IF NOT EXISTS quaycontainerpush (
	name          String,
	repository    String,
	nameSpace     String,
	dockerURL     String,
	homePage      String,
	tag           String,
	Event         String
) engine=File(TabSeparated);

-- trivySbomTable
CREATE TABLE IF NOT EXISTS trivysbom (
	id                    UUID,
	schema                String,
	bom_format            String,
	spec_version          String,
	serial_number         String,
	version               INTEGER,
	metadata_timestamp    DateTime('UTC'),
	metatool_vendor       String,
	metatool_name         String,
	metatool_version      String,
	component_bom_ref     String,
	component_type        String,
	component_name        String,
	component_version     String,
	component_property_name String,
	component_property_value String,
	component_hash_alg    String,
	component_hash_content String,
	component_license_exp String,
	component_purl        String,
	dependency_ref        String
) engine=File(TabSeparated);

-- AzureDevopsTable
CREATE TABLE IF NOT EXISTS azure_devops (
	Author       String,
	Provider     String,
	CommitID     String,
	CommitUrl    String,
	EventType    String,
	RepoName     String,
	TimeStamp    String,
	Event        String
) engine=File(TabSeparated);

-- GithubTable
CREATE TABLE IF NOT EXISTS github (
	Author       String,
	Provider     String,
	CommitID     String,
	CommitUrl    String,
	EventType    String,
	RepoName     String,
	TimeStamp    String,
	Event        String
) engine=File(TabSeparated);

-- GitlabTable
CREATE TABLE IF NOT EXISTS gitlab (
	Author       String,
	Provider     String,
	CommitID     String,
	CommitUrl    String,
	EventType    String,
	RepoName     String,
	TimeStamp    String,
	Event        String
) engine=File(TabSeparated);

-- BitbucketTable
CREATE TABLE IF NOT EXISTS bitbucket (
	Author       String,
	Provider     String,
	CommitID     String,
	CommitUrl    String,
	EventType    String,
	RepoName     String,
	TimeStamp    String,
	Event        String
) engine=File(TabSeparated);

-- GiteaTable
CREATE TABLE IF NOT EXISTS gitea (
	Author       String,
	Provider     String,
	CommitID     String,
	CommitUrl    String,
	EventType    String,
	RepoName     String,
	TimeStamp    String,
	Event        String
) engine=File(TabSeparated);
