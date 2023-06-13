package models

import (
	"fmt"
	"io"
	"strconv"
	"time"
)

type Connection interface {
	IsConnection()
	// Total count of items
	GetTotalCount() int
	// Information to aid in pagination of list
	GetPageInfo() *PageInfo
}

type InstallSettings interface {
	IsInstallSettings()
}

type Node interface {
	IsNode()
	// A globally-unique identifier
	GetID() string
	// The date and time (ISO 8601 format) when the node was created
	GetCreatedAt() time.Time
	// The date and time (ISO 8601 format) when the node was last updated
	GetUpdatedAt() time.Time
}

type AWSSettingsInput struct {
	Region     AWSRegion `json:"region"`
	AccountID  string    `json:"accountId"`
	IamRoleArn string    `json:"iamRoleArn"`
}

type AccountConnection struct {
	IamRole   string    `json:"IamRole"`
	Status    string    `json:"Status"`
	UpdatedAt time.Time `json:"UpdatedAt"`
}

// An auto-generated type for paginating through multiple Apps
type AppConnection struct {
	TotalCount int       `json:"totalCount"`
	PageInfo   *PageInfo `json:"pageInfo"`
	// A list of edges
	Edges []*AppEdge `json:"edges"`
}

func (AppConnection) IsConnection() {}

// Total count of items
//
//nolint:all
func (this AppConnection) GetTotalCount() int { return this.TotalCount }

// Information to aid in pagination of list
//
//nolint:all
func (this AppConnection) GetPageInfo() *PageInfo { return this.PageInfo }

// An auto-generated type which holds one App and a cursor during pagination
type AppEdge struct {
	// A cursor for use in pagination
	Cursor string `json:"cursor"`
	// The item at the end of AppEdge
	Node *App `json:"node"`
}

type AppInput struct {
	ID          *string `json:"id"`
	Name        string  `json:"name"`
	OrgID       string  `json:"orgId"`
	CreatedByID *string `json:"createdById"`

	// OverrideID is used to override the id when creating an object, and should only be used locally.
	OverrideID *string `json:"overrideId"`
}

// An auto-generated type for paginating through multiple Components
type ComponentConnection struct {
	TotalCount int       `json:"totalCount"`
	PageInfo   *PageInfo `json:"pageInfo"`
	// A list of edges
	Edges []*ComponentEdge `json:"edges"`
}

func (ComponentConnection) IsConnection() {}

// Total count of items
//
//nolint:all
func (this ComponentConnection) GetTotalCount() int { return this.TotalCount }

// Information to aid in pagination of list
//
//nolint:all
func (this ComponentConnection) GetPageInfo() *PageInfo { return this.PageInfo }

// An auto-generated type which holds one Component and a cursor during pagination
type ComponentEdge struct {
	// A cursor for use in pagination
	Cursor string `json:"cursor"`
	// The item at the end of ComponentEdge
	Node *Component `json:"node"`
}

type ComponentInput struct {
	AppID       string  `json:"appId"`
	ID          *string `json:"id"`
	Name        string  `json:"name"`
	CreatedByID string  `json:"created_by_id"`
	Config      []byte  `json:"component_config"`
}

// Filters returned Connection results
type ConnectionFilter struct {
	CreatedAt *DateOperators `json:"createdAt"`
	UpdatedAt *DateOperators `json:"updatedAt"`
}

type ConnectionOptions struct {
	// Returns the elements that come after the specified cursor
	After *string `json:"after"`
	// Returns the elements that come before the specified cursor
	Before *string `json:"before"`
	// Returns first (n) elements
	First *int `json:"first"`
	// Returns last (n) elements
	Last *int `json:"last"`
	// Filter underlying list
	Filter *ConnectionFilter `json:"filter"`
	// Reverse the order of the underlying list
	Reverse *bool `json:"reverse"`
	// Sort the underlying list by the given key
	SortKey *OrderDirection `json:"sortKey"`
	// Limit: the total number of results returned
	Limit *int `json:"limit"`
}

// Operator to filter on DateTime field
type DateOperators struct {
	// Filter by exact date
	Eq *time.Time `json:"eq"`
	// Filter by before a date
	Before *time.Time `json:"before"`
	// Filter by after a date
	After *time.Time `json:"after"`
	// Filter by between two dates
	Between *DateRange `json:"between"`
}

// A range of time between two dates
type DateRange struct {
	Begin *time.Time `json:"begin"`
	End   *time.Time `json:"end"`
}

// An auto-generated type for paginating through multiple Deployments
type DeploymentConnection struct {
	TotalCount int       `json:"totalCount"`
	PageInfo   *PageInfo `json:"pageInfo"`
	// A list of edges
	Edges []*DeploymentEdge `json:"edges"`
}

func (DeploymentConnection) IsConnection() {}

// Total count of items
//
//nolint:all
func (this DeploymentConnection) GetTotalCount() int { return this.TotalCount }

// Information to aid in pagination of list
//
//nolint:all
func (this DeploymentConnection) GetPageInfo() *PageInfo { return this.PageInfo }

// An auto-generated type which holds one Deployment and a cursor during pagination
type DeploymentEdge struct {
	// A cursor for use in pagination
	Cursor string `json:"cursor"`
	// The item at the end of DeploymentEdge
	Node *Deployment `json:"node"`
}

type DeploymentInput struct {
	ComponentID string  `json:"componentId"`
	CreatedByID *string `json:"createdById"`
}

type GCPSettingsInput struct {
	Bogus string `json:"bogus"`
}

// An auto-generated type for paginating through multiple Installs
type InstallConnection struct {
	TotalCount int       `json:"totalCount"`
	PageInfo   *PageInfo `json:"pageInfo"`
	// A list of edges
	Edges []*InstallEdge `json:"edges"`
}

func (InstallConnection) IsConnection() {}

// Total count of items
//
//nolint:all
func (this InstallConnection) GetTotalCount() int { return this.TotalCount }

// Information to aid in pagination of list
//
//nolint:all
func (this InstallConnection) GetPageInfo() *PageInfo { return this.PageInfo }

// An auto-generated type which holds one Install and a cursor during pagination
type InstallEdge struct {
	// A cursor for use in pagination
	Cursor string `json:"cursor"`
	// The item at the end of InstallEdge
	Node *Install `json:"node"`
}

type InstallInput struct {
	ID          *string           `json:"id"`
	Name        string            `json:"name"`
	AppID       string            `json:"appId"`
	CreatedByID *string           `json:"createdById"`
	AwsSettings *AWSSettingsInput `json:"awsSettings"`
	GcpSettings *GCPSettingsInput `json:"gcpSettings"`
	OverrideID  *string           `json:"overrideId"`
}

// An auto-generated type for paginating through multiple Orgs
type OrgConnection struct {
	TotalCount int       `json:"totalCount"`
	PageInfo   *PageInfo `json:"pageInfo"`
	// A list of edges
	Edges []*OrgEdge `json:"edges"`
}

func (OrgConnection) IsConnection() {}

// Total count of items
//
//nolint:all
func (this OrgConnection) GetTotalCount() int { return this.TotalCount }

// Information to aid in pagination of list
//
//nolint:all
func (this OrgConnection) GetPageInfo() *PageInfo { return this.PageInfo }

// An auto-generated type which holds one Org and a cursor during pagination
type OrgEdge struct {
	// A cursor for use in pagination
	Cursor string `json:"cursor"`
	// The item at the end of OrgEdge
	Node *Org `json:"node"`
}

type OrgInput struct {
	ID              *string `json:"id"`
	Name            string  `json:"name"`
	OwnerID         string  `json:"ownerId"`
	GithubInstallID *string `json:"githubInstallId"`

	// OverrideID is used to override the id when creating an object, and should only be used locally.
	OverrideID *string `json:"overrideId"`
}

// Returns information about pagination in a connection, in accordance with the Relay specification
type PageInfo struct {
	// The cursor corresponding to the last node in edges
	EndCursor *string `json:"endCursor"`
	// Whether there are more pages to fetch following the current page
	HasNextPage bool `json:"hasNextPage"`
	// Whether there are any pages prior to the current page
	HasPreviousPage bool `json:"hasPreviousPage"`
	// The cursor corresponding to the first node in edges
	StartCursor *string `json:"startCursor"`
}

type Repo struct {
	DefaultBranch *string `json:"defaultBranch"`
	FullName      *string `json:"fullName"`
	Name          *string `json:"name"`
	Owner         *string `json:"owner"`
	Private       *bool   `json:"private"`
	URL           *string `json:"url"`
}

// An auto-generated type for paginating through multiple Installs
type RepoConnection struct {
	TotalCount int `json:"totalCount"`
	// A list of edges
	Edges []*RepoEdge `json:"edges"`
}

// An auto-generated type which holds one Install and a cursor during pagination
type RepoEdge struct {
	// The item at the end of RepoEdge
	Node *Repo `json:"node"`
}

// Operator to filter on String field
type StringOperators struct {
	Eq          *string  `json:"eq"`
	NotEq       *string  `json:"notEq"`
	Contains    *string  `json:"contains"`
	NotContains *string  `json:"notContains"`
	In          []string `json:"in"`
	NotIn       []string `json:"notIn"`
	Regex       *string  `json:"regex"`
}

type UserOrgInput struct {
	UserID string `json:"userId"`
	OrgID  string `json:"orgId"`
}

type SandboxVersionInput struct {
	ID             string `json:"id"`
	SandboxName    string `json:"sandboxName"`
	SandboxVersion string `json:"sandboxVersion"`
	TfVersion      string `json:"tfVersion"`
}

type AWSRegion string

const (
	AWSRegionUsEast1 AWSRegion = "US_EAST_1"
	AWSRegionUsEast2 AWSRegion = "US_EAST_2"
	AWSRegionUsWest1 AWSRegion = "US_WEST_1"
	AWSRegionUsWest2 AWSRegion = "US_WEST_2"
)

var AllAWSRegion = []AWSRegion{
	AWSRegionUsEast1,
	AWSRegionUsEast2,
	AWSRegionUsWest1,
	AWSRegionUsWest2,
}

//nolint:all
func (e AWSRegion) IsValid() bool {
	switch e {
	case AWSRegionUsEast1, AWSRegionUsEast2, AWSRegionUsWest1, AWSRegionUsWest2:
		return true
	}
	return false
}

func (e AWSRegion) String() string {
	return string(e)
}

func (e *AWSRegion) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = AWSRegion(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid AWSRegion", str)
	}
	return nil
}

func (e AWSRegion) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

// Represents a collection of general settings and information about a piece of a App
type OrderDirection string

const (
	// Sort accending 'A-Z'
	OrderDirectionAsc OrderDirection = "ASC"
	// Sort decending 'Z-A'
	OrderDirectionDesc OrderDirection = "DESC"
)

var AllOrderDirection = []OrderDirection{
	OrderDirectionAsc,
	OrderDirectionDesc,
}

func (e OrderDirection) IsValid() bool {
	switch e {
	case OrderDirectionAsc, OrderDirectionDesc:
		return true
	}
	return false
}

func (e OrderDirection) String() string {
	return string(e)
}

func (e *OrderDirection) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = OrderDirection(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid OrderDirection", str)
	}
	return nil
}

func (e OrderDirection) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
