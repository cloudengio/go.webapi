# Package [cloudeng.io/webapi/clients/papersapp/papersappsdk](https://pkg.go.dev/cloudeng.io/webapi/clients/papersapp/papersappsdk?tab=doc)

```go
import cloudeng.io/webapi/clients/papersapp/papersappsdk
```


## Types
### Type ArticleMetadata
```go
type ArticleMetadata struct {

	// Abstract text
	Abstract string `json:"abstract,omitempty"`

	// Authors names
	Authors []string `json:"authors"`

	// Chapter title (books only)
	Chapter string `json:"chapter,omitempty"`

	// eisbn
	Eisbn string `json:"eisbn,omitempty"`

	// eissn
	Eissn string `json:"eissn,omitempty"`

	// isbn
	Isbn string `json:"isbn,omitempty"`

	// Journal ISSN (e.g. 1523-4681)
	Issn string `json:"issn,omitempty"`

	// Issue number (e.g. 12)
	Issue string `json:"issue,omitempty"`

	// Journal name (e.g. Journal of Bone and Mineral Research)
	Journal string `json:"journal,omitempty"`

	// Abbreviation of journal title (e.g. J Bone Miner Res)
	JournalAbbrev string `json:"journal_abbrev,omitempty"`

	// pagination
	Pagination string `json:"pagination,omitempty"`

	// Article or book title (e.g. Parathyroid hormone: past and present)
	Title string `json:"title,omitempty"`

	// Publisher URL (e.g. https://asbmr.onlinelibrary.wiley.com/doi/10.1002/jbmr.178)
	URL string `json:"url,omitempty"`

	// Volume number (e.g. 1)
	Volume string `json:"volume,omitempty"`

	// Publication year (e.g. 1523-4681)
	// NOTE: swagger thinks this is a float, most likely because of the example
	// above.
	Year int `json:"year,omitempty"`
}
```
ArticleMetadata Article Metadata Schema

swagger:model ArticleMetadata


### Type Collection
```go
type Collection struct {

	// custom fields
	CustomFields []*CustomField `json:"custom_fields"`

	// Unique collection id (e.g. 7f5e3fe2-bcd7-42b5-972e-c271f5449977)
	ID string `json:"id,omitempty"`

	// Collection name (e.g. Elephant Shark Shared Library)
	Name string `json:"name,omitempty"`

	// owner
	Owner *CollectionOwner `json:"owner,omitempty"`

	// Indicates if the collection is shared or personal
	Shared bool `json:"shared,omitempty"`
}
```
Collection Collection Schema

swagger:model Collection


### Type CollectionOwner
```go
type CollectionOwner struct {

	// email
	Email string `json:"email,omitempty"`

	// User unique ID
	ID string `json:"id,omitempty"`

	// name
	Name string `json:"name,omitempty"`
}
```
CollectionOwner Owner of collection (can be null)

swagger:model CollectionOwner


### Type Collections
```go
type Collections struct {

	// collections
	Collections []*Collection `json:"collections"`
}
```
Collections Collections Schema

# Set of collections

swagger:model Collections


### Type CustomField
```go
type CustomField struct {

	// display
	Display string `json:"display,omitempty"`

	// field
	Field string `json:"field,omitempty"`

	// show in details
	ShowInDetails bool `json:"show_in_details,omitempty"`

	// show in table
	ShowInTable bool `json:"show_in_table,omitempty"`

	// type
	Type string `json:"type,omitempty"`
}
```
CustomField custom field

swagger:model CustomField


### Type ExtIds
```go
type ExtIds struct {

	// arxiv
	Arxiv string `json:"arxiv,omitempty"`

	// Item DOI (e.g. 10.1002/jbmr.178)
	Doi string `json:"doi,omitempty"`

	// gsid
	Gsid string `json:"gsid,omitempty"`

	// Item patent ID (e.g. CH-708619-A1)
	PatentID string `json:"patent_id,omitempty"`

	// pmcid
	Pmcid string `json:"pmcid,omitempty"`

	// Item PMID (e.g. 20614475)
	Pmid string `json:"pmid,omitempty"`
}
```
ExtIds ExtIds schema

swagger:model ExtIds


### Type File
```go
type File struct {

	// File creation timestamp
	Created string `json:"created,omitempty"`

	// File format
	FileType string `json:"file_type,omitempty"`

	// File name
	Name string `json:"name,omitempty"`

	// Number of pages
	Pages float64 `json:"pages,omitempty"`

	// SHA256 hash of the file
	Sha256 string `json:"sha256,omitempty"`

	// File size in bytes
	Size float64 `json:"size,omitempty"`

	// File type
	// Enum: [article supplement]
	Type string `json:"type,omitempty"`

	// File download url
	URL string `json:"url,omitempty"`
}
```
File File Schema

swagger:model File


### Type ImportData
```go
type ImportData struct {

	// Client that did the import
	ImportedBy string `json:"imported_by,omitempty"`

	// original id
	OriginalID string `json:"original_id,omitempty"`

	// original type
	OriginalType string `json:"original_type,omitempty"`

	// Source from which the item is imported
	Source string `json:"source,omitempty"`
}
```
ImportData import data

swagger:model ImportData


### Type Item
```go
type Item struct {

	// article
	Article *ArticleMetadata `json:"article,omitempty"`

	// custom metadata
	CustomMetadata any `json:"custom_metadata,omitempty"`

	// custom type
	CustomType string `json:"custom_type,omitempty"`

	// ext ids
	ExtIds *ExtIds `json:"ext_ids,omitempty"` //nolint:revive

	// files
	Files []File `json:"files,omitempty"`

	// Unique item id (e.g. 3cdfbfb9-07c2-4924-9eb3-3b790df74c0e)
	ID string `json:"id,omitempty"`

	// import data
	ImportData *ImportData `json:"import_data,omitempty"`

	// item type
	ItemType ItemType `json:"item_type,omitempty"`

	// SHA256 hash of the primary item PDF file
	PdfHash string `json:"pdf_hash,omitempty"`

	// user data
	UserData *UserData `json:"user_data,omitempty"`
}
```
Item Item Response Schema

swagger:model Item


### Type ItemType
```go
type ItemType string
```
ItemType ItemType schema

swagger:model ItemType

### Constants
### ItemTypeAbstract, ItemTypeArticle, ItemTypeArtwork, ItemTypeAudioRecording, ItemTypeBill, ItemTypeBlogPost, ItemTypeBook, ItemTypeBookSection, ItemTypeCase, ItemTypeClinicalTrial, ItemTypeComputerProgram, ItemTypeConferencePaper, ItemTypeDictionaryEntry, ItemTypeDocument, ItemTypeEmail, ItemTypeEncyclopediaArticle, ItemTypeFilm, ItemTypeForumPost, ItemTypeGuidanceDocuments, ItemTypeHearing, ItemTypeInstantMessage, ItemTypeInterview, ItemTypeLetter, ItemTypeMagazine, ItemTypeManuscript, ItemTypeMap, ItemTypeNewspaperArticle, ItemTypePatent, ItemTypePodcast, ItemTypePosterPresentation, ItemTypePresentation, ItemTypeRadioBroadcast, ItemTypeReport, ItemTypeStatute, ItemTypeThesis, ItemTypeTvBroadcast, ItemTypeVideoRecording, ItemTypeWebpage
```go
// ItemTypeAbstract captures enum value "abstract"
ItemTypeAbstract ItemType = "abstract"
// ItemTypeArticle captures enum value "article"
ItemTypeArticle ItemType = "article"
// ItemTypeArtwork captures enum value "artwork"
ItemTypeArtwork ItemType = "artwork"
// ItemTypeAudioRecording captures enum value "audio_recording"
ItemTypeAudioRecording ItemType = "audio_recording"
// ItemTypeBill captures enum value "bill"
ItemTypeBill ItemType = "bill"
// ItemTypeBlogPost captures enum value "blog_post"
ItemTypeBlogPost ItemType = "blog_post"
// ItemTypeBook captures enum value "book"
ItemTypeBook ItemType = "book"
// ItemTypeBookSection captures enum value "book_section"
ItemTypeBookSection ItemType = "book_section"
// ItemTypeCase captures enum value "case"
ItemTypeCase ItemType = "case"
// ItemTypeClinicalTrial captures enum value "clinical_trial"
ItemTypeClinicalTrial ItemType = "clinical_trial"
// ItemTypeComputerProgram captures enum value "computer_program"
ItemTypeComputerProgram ItemType = "computer_program"
// ItemTypeConferencePaper captures enum value "conference_paper"
ItemTypeConferencePaper ItemType = "conference_paper"
// ItemTypeDictionaryEntry captures enum value "dictionary_entry"
ItemTypeDictionaryEntry ItemType = "dictionary_entry"
// ItemTypeDocument captures enum value "document"
ItemTypeDocument ItemType = "document"
// ItemTypeEmail captures enum value "email"
ItemTypeEmail ItemType = "email"
// ItemTypeEncyclopediaArticle captures enum value "encyclopedia_article"
ItemTypeEncyclopediaArticle ItemType = "encyclopedia_article"
// ItemTypeFilm captures enum value "film"
ItemTypeFilm ItemType = "film"
// ItemTypeForumPost captures enum value "forum_post"
ItemTypeForumPost ItemType = "forum_post"
// ItemTypeGuidanceDocuments captures enum value "guidance_documents"
ItemTypeGuidanceDocuments ItemType = "guidance_documents"
// ItemTypeHearing captures enum value "hearing"
ItemTypeHearing ItemType = "hearing"
// ItemTypeInstantMessage captures enum value "instant_message"
ItemTypeInstantMessage ItemType = "instant_message"
// ItemTypeInterview captures enum value "interview"
ItemTypeInterview ItemType = "interview"
// ItemTypeLetter captures enum value "letter"
ItemTypeLetter ItemType = "letter"
// ItemTypeMagazine captures enum value "magazine"
ItemTypeMagazine ItemType = "magazine"
// ItemTypeManuscript captures enum value "manuscript"
ItemTypeManuscript ItemType = "manuscript"
// ItemTypeMap captures enum value "map"
ItemTypeMap ItemType = "map"
// ItemTypeNewspaperArticle captures enum value "newspaper_article"
ItemTypeNewspaperArticle ItemType = "newspaper_article"
// ItemTypePatent captures enum value "patent"
ItemTypePatent ItemType = "patent"
// ItemTypePodcast captures enum value "podcast"
ItemTypePodcast ItemType = "podcast"
// ItemTypePosterPresentation captures enum value "poster_presentation"
ItemTypePosterPresentation ItemType = "poster_presentation"
// ItemTypePresentation captures enum value "presentation"
ItemTypePresentation ItemType = "presentation"
// ItemTypeRadioBroadcast captures enum value "radio_broadcast"
ItemTypeRadioBroadcast ItemType = "radio_broadcast"
// ItemTypeReport captures enum value "report"
ItemTypeReport ItemType = "report"
// ItemTypeStatute captures enum value "statute"
ItemTypeStatute ItemType = "statute"
// ItemTypeThesis captures enum value "thesis"
ItemTypeThesis ItemType = "thesis"
// ItemTypeTvBroadcast captures enum value "tv_broadcast"
ItemTypeTvBroadcast ItemType = "tv_broadcast"
// ItemTypeVideoRecording captures enum value "video_recording"
ItemTypeVideoRecording ItemType = "video_recording"
// ItemTypeWebpage captures enum value "webpage"
ItemTypeWebpage ItemType = "webpage"

```




### Type Items
```go
type Items struct {

	// Items list
	Items []*Item `json:"items"`

	Total int64 `json:"total,omitempty"`

	// Scroll id to be used to fetch the next batch of results
	ScrollID string `json:"scroll_id,omitempty"`
}
```
Items Items Response Schema

# Set of items

swagger:model Items


### Type List
```go
type List struct {

	// Indicates whether the list was deleted or not
	Deleted bool `json:"deleted,omitempty"`

	// List id (e.g. 61a19540-77b8-46b2-9b82-5ca8a3af88d7)
	ID string `json:"id,omitempty"`

	// List modification timestamp (e.g. 2020-04-29T16:12:17Z)
	Modified string `json:"modified,omitempty"`

	// List name (e.g. Shark Folder)
	Name string `json:"name,omitempty"`

	// List parent id
	ParentID string `json:"parent_id,omitempty"`
}
```
List List Schema

swagger:model List

### Methods

```go
func (m *List) ContextValidate(ctx context.Context, formats strfmt.Registry) error
```
ContextValidate validates this list based on context it is used


```go
func (m *List) MarshalBinary() ([]byte, error)
```
MarshalBinary interface implementation


```go
func (m *List) UnmarshalBinary(b []byte) error
```
UnmarshalBinary interface implementation


```go
func (m *List) Validate(formats strfmt.Registry) error
```
Validate validates this list




### Type Lists
```go
type Lists struct {

	// Lists list
	Lists []*List `json:"lists"`
}
```
Lists Lists Schema

swagger:model Lists

### Methods

```go
func (m *Lists) ContextValidate(ctx context.Context, formats strfmt.Registry) error
```
ContextValidate validate this lists based on the context it is used


```go
func (m *Lists) MarshalBinary() ([]byte, error)
```
MarshalBinary interface implementation


```go
func (m *Lists) UnmarshalBinary(b []byte) error
```
UnmarshalBinary interface implementation


```go
func (m *Lists) Validate(formats strfmt.Registry) error
```
Validate validates this lists




### Type Token
```go
type Token struct {

	// token
	Token string `json:"token,omitempty"`
}
```
Token token

swagger:model Token


### Type UserData
```go
type UserData struct {

	// Hex color (e.g. "#fe3018")
	Color string `json:"color,omitempty"`

	// Item creation timestamp
	Created string `json:"created,omitempty"`

	// User note
	Notes string `json:"notes,omitempty"`

	// Item flag/star
	Star bool `json:"star,omitempty"`

	// Item hashtags/keywords (e.g. Shark)
	Tags []string `json:"tags"`
}
```
UserData UserData Schema

swagger:model UserData





