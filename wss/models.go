package wss

type AsyncProcess struct {
	Uuid               string `json:"uuid" url:"uuid,omitempty"`
	RequestToken       string `json:"requestToken" url:"requestToken,omitempty"`
	ContextId          string `json:"contextId" url:"contextId,omitempty"`
	ContextType        string `json:"contextType" url:"contextType,omitempty"`
	ProcessType        string `json:"processType" url:"processType,omitempty"`
	UserEmail          string `json:"userEmail" url:"userEmail,omitempty"`
	MessageContentSha1 string `json:"messageContentSha1" url:"messageContentSha1,omitempty"`
	Status             string `json:"status" url:"status,omitempty"`
	Created            string `json:"created" url:"created,omitempty"`
	Modified           string `json:"modified" url:"modified,omitempty"`
}

type ProcessStatusResponse struct {
	AsyncProcessStatus AsyncProcess `json:"asyncProcessStatus" url:"asyncProcessStatus,omitempty"`
}

type WhiteSourceEnv struct {
	ApiKey       string `yaml:"apiKey"`
	UserKey      string `yaml:"userKey"`
	ProjectName  string `yaml:"projectName"`
	ProductName  string `yaml:"productName"`
	ProductToken string `yaml:"productToken"`
	WSSUrl       string `yaml:"wss.url"`
	Offline      string `yaml:"offline"`
}

type CheckSums struct {
	SHA1                string `json:"SHA1" url:"SHA1,omitempty"`
	SHA1_OTHER_PLATFORM string `json:"SHA1_OTHER_PLATFORM" url:"SHA1_OTHER_PLATFORM,omitempty"`
}

type Dependency struct {
	ArtifactId        string    `json:"artifactId" url:"artifactId,omitempty"`
	Sha1              string    `json:"sha1" url:"sha1,omitempty"`
	OtherPlatformSha1 string    `json:"otherPlatformSha1" url:"otherPlatformSha1,omitempty"`
	SystemPath        string    `json:"systemPath" url:"systemPath,omitempty"`
	Optional          bool      `json:"optional" url:"optional,omitempty"`
	Filename          string    `json:"filename" url:"filename,omitempty"`
	Checksums         CheckSums `json:"checksums" url:"checksums,omitempty"`
	Deduped           bool      `json:"deduped" url:"deduped,omitempty"`
}

type Coordinates struct {
	ArtifactId string `json:"artifactId" url:"artifactId,omitempty"`
	Version    string `json:"version" url:"version,omitempty"`
}

type Project struct {
	Coordinates  Coordinates  `json:"coordinates" url:"coordinates,omitempty"`
	Dependencies []Dependency `json:"dependencies" url:"dependencies,omitempty"`
}

type ExtraProperties struct {
	ContributionsAvailable string `json:"contributionsAvailable" url:"contributionsAvailable,omitempty"`
}

type StepsSummaryInfo struct {
	StepName             string `json:"stepName" url:"stepName,omitempty"`
	TotalElapsedTime     int    `json:"totalElapsedTime" url:"totalElapsedTime,omitempty"`
	IsSubStep            bool   `json:"isSubStep" url:"isSubStep,omitempty"`
	StepCompletionStatus string `json:"stepCompletionStatus" url:"stepCompletionStatus,omitempty"`
}

type ScanMethod struct {
	Type    string `json:"type" url:"type,omitempty"`
	Version string `json:"version" url:"version,omitempty"`
}

type ScanSummaryInfo struct {
	TotalElapsedTime int                `json:"totalElapsedTime" url:"totalElapsedTime,omitempty"`
	StepsSummaryInfo []StepsSummaryInfo `json:"stepsSummaryInfo" url:"stepsSummaryInfo,omitempty"`
	IsPrivileged     bool               `json:"isPrivileged" url:"isPrivileged,omitempty"`
	ScanMethod       ScanMethod         `json:"scanMethod" url:"scanMethod,omitempty"`
}

type UpdateRequestOriginal struct {
	UpdateType              string          `json:"updateType" url:"updateType,omitempty"`
	Type                    string          `json:"type" url:"type,omitempty"`
	Agent                   string          `json:"agent" url:"agent,omitempty"`
	AgentVersion            string          `json:"agentVersion" url:"agentVersion,omitempty"`
	Token                   string          `json:"orgToken" url:"token,omitempty"`
	UserKey                 string          `json:"userKey" url:"userKey,omitempty"`
	Product                 string          `json:"product" url:"product,omitempty"`
	TimeStamp               int             `json:"timeStamp" url:"timeStamp,omitempty"`
	Diff                    []Project       `json:"projects" url:"diff,omitempty"`
	AggregateModules        bool            `json:"aggregateModules" url:"aggregateModules,omitempty"`
	PreserveModuleStructure bool            `json:"preserveModuleStructure" url:"preserveModuleStructure,omitempty"`
	ProductToken            string          `json:"productToken" url:"productToken,omitempty"`
	ExtraProperties         ExtraProperties `json:"extraProperties" url:"extraProperties,omitempty"`
	ScanSummaryInfo         ScanSummaryInfo `json:"scanSummaryInfo" url:"scanSummaryInfo,omitempty"`
}

type UpdateRequestReq struct {
	UpdateType   string    `json:"updateType" url:"updateType,omitempty"`
	Type         string    `json:"type" url:"type,omitempty"`
	Agent        string    `json:"agent" url:"agent,omitempty"`
	AgentVersion string    `json:"agentVersion" url:"agentVersion,omitempty"`
	Token        string    `json:"token" url:"token,omitempty"`
	UserKey      string    `json:"userKey" url:"userKey,omitempty"`
	Product      string    `json:"product" url:"product,omitempty"`
	TimeStamp    int       `json:"timeStamp" url:"timeStamp,omitempty"`
	Diff         []Project `json:"diff" url:"diff,omitempty"`
}

type UploadResponseStatus struct {
	EnvelopeVersion string `json:"envelopeVersion" url:"envelopeVersion,omitempty"`
	Status          int    `json:"status" url:"status,omitempty"`
	Message         string `json:"message" url:"message,omitempty"`
	Data            string `json:"data" url:"data,omitempty"`
	RequestToken    string `json:"requestToken" url:"requestToken,omitempty"`
}

type ProjectInfo struct {
	ProjectName  string `json:"projectName" url:"projectName,omitempty"`
	ProjectId    int    `json:"projectId" url:"projectId,omitempty"`
	ProjectToken string `json:"projectToken" url:"projectToken,omitempty"`
}

type UploadResponseData struct {
	RemoveIfExist         bool                   `json:"removeIfExist" url:"removeIfExist,omitempty"`
	UpdatedProjects       []string               `json:"updatedProjects" url:"updatedProjects,omitempty"`
	CreatedProjects       []string               `json:"createdProjects" url:"createdProjects,omitempty"`
	ProjectNamesToIds     map[string]int         `json:"projectNamesToIds" url:"projectNamesToIds,omitempty"`
	ProjectNamesToDetails map[string]ProjectInfo `json:"projectNamesToDetails" url:"projectNamesToDetails,omitempty"`
	Organization          string                 `json:"organization" url:"organization,omitempty"`
	RequestToken          string                 `json:"requestToken" url:"requestToken,omitempty"`
}

type GenerateProjectReportAsyncRequest struct {
	RequestType  string `json:"requestType" url:"requestType,omitempty"`
	ProjectToken string `json:"projectToken" url:"projectToken,omitempty"`
	UserKey      string `json:"userKey" url:"userKey,omitempty"`
	ReportType   string `json:"reportType" url:"reportType,omitempty"`
	Format       string `json:"format" url:"format,omitempty"`
}

type AsyncProcessStatusRequest struct {
	RequestType string `json:"requestType" url:"requestType,omitempty"`
	UserKey     string `json:"userKey" url:"userKey,omitempty"`
	OrgToken    string `json:"orgToken" url:"orgToken,omitempty"`
	Uuid        string `json:"uuid" url:"uuid,ommitempty"`
}

// type ProjectAlertRequest struct {
// 	RequestType  string `json:"requestType" url:"requestType,omitempty"`
// 	UserKey      string `json:"userKey" url:"userKey,omitempty"`
// 	ProjectToken string `json:"projectToken" url:"projectToken,omitempty"`
// }

type ProjectInfoRequest struct {
	RequestType  string `json:"requestType" url:"requestType,omitempty"`
	UserKey      string `json:"userKey" url:"userKey,omitempty"`
	ProjectToken string `json:"projectToken" url:"projectToken,omitempty"`
}

type ProjectInventoryRequest struct {
	RequestType        string   `json:"requestType" url:"requestType,omitempty"`
	ProjectToken       string   `json:"projectToken" url:"projectToken,omitempty"`
	UserKey            string   `json:"userKey" url:"userKey,omitempty"`
	Format             string   `json:"format" url:"format,omitempty"`
	ExtraLibraryFields []string `json:"extraLibraryFields" url:"extraLibraryFields,omitempty"`
}

type ProjectRiskRequest struct {
	RequestType  string `json:"requestType" url:"requestType,omitempty"`
	UserKey      string `json:"userKey" url:"userKey,omitempty"`
	ProjectToken string `json:"projectToken" url:"projectToken,omitempty"`
}

type ProfileInfo struct {
	CopyrightRiskScore string `json:"copyrightRiskScore" url:"copyrightRiskScore,omitempty"`
	PatentRiskScore    string `json:"patentRiskScore" url:"patentRiskScore,omitempty"`
	Copyleft           string `json:"copyleft" url:"copyleft,omitempty"`
	RoyaltyFree        string `json:"royaltyFree" url:"royaltyFree,omitempty"`
	Linking            string `json:"linking" url:"linking,omitempty"`
}

type License struct {
	Name          string      `json:"name" url:"name,omitempty"`
	Url           string      `json:"url" url:"url,omitempty"`
	ProfileInfo   ProfileInfo `json:"profileInfo" url:"profileInfo,omitempty"`
	References    string      `json:"references" url:"references,omitempty"`
	ReferenceType string      `json:"referenceType" url:"referenceType,omitempty"`
}

type Reference struct {
	Url                 string `json:"url" url:"url,omitempty"`
	HomePage            string `json:"homePage" url:"homePage,omitempty"`
	GenericPackageIndex string `json:"genericPackageIndex" url:"genericPackageIndex,omitempty"`
}

type OutdatedModel struct {
	OutdatedLibraryDate string `json:"outdatedLibraryDate" url:"outdatedLibraryDate,omitempty"`
	NewestVersion       string `json:"newestVersion" url:"newestVersion,omitempty"`
	NewestLibraryDate   string `json:"newestLibraryDate" url:"newestLibraryDate,omitempty"`
	VersionsInBetween   int    `json:"versionsInBetween" url:"versionsInBetween,omitempty"`
}

type Library struct {
	KeyUuid      string        `json:"keyUuid" url:"keyUuid,omitempty"`
	KeyId        int           `json:"keyId" url:"keyId,omitempty"`
	Type         string        `json:"type" url:"type,omitempty"`
	Languages    string        `json:"languages" url:"languages,omitempty"`
	References   Reference     `json:"references" url:"references,omitempty"`
	Outdated     bool          `json:"outdated" url:"outdated,omitempty"`
	MatchType    string        `json:"matchType" url:"matchType,omitempty"`
	OtdatedModel OutdatedModel `json:"outdatedModel" url:"outdatedModel,omitempty"`
	Sha1         string        `json:"sha1" url:"sha1,omitempty"`
	Name         string        `json:"name" url:"name,omitempty"`
	ArtifactId   string        `json:"artifactId" url:"artifactId,omitempty"`
	Version      string        `json:"version" url:"version,omitempty"`
	GroupId      string        `json:"groupId" url:"groupId,omitempty"`
	Licenses     []License     `json:"licenses" url:"licenses,omitempty"`
}

type ProjectVitals struct {
	ProductName     string `json:"productName" url:"productName,omitempty"`
	Name            string `json:"name" url:"name,omitempty"`
	Token           string `json:"token" url:"token,omitempty"`
	CreationDate    string `json:"creationDate" url:"creationDate,omitempty"`
	LastUpdatedDate string `json:"lastUpdatedDate" url:"lastUpdatedDate,omitempty"`
}

type ProjectScanInfo struct {
	ProjectVitals ProjectVitals `json:"projectVitals" url:"projectVitlas,omitempty"`
}
