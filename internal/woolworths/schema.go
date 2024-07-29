package woolworths

import (
	"time"

	"github.com/shopspring/decimal"
)

type JSONTime struct {
	time.Time
}

type departmentInfo struct {
	NodeID              departmentID `json:"NodeId"`
	Description         string       `json:"Description"`
	NodeLevel           int          `json:"NodeLevel"`
	ParentNodeID        *string      `json:"ParentNodeId"`
	DisplayOrder        int          `json:"DisplayOrder"`
	IsRestricted        bool         `json:"IsRestricted"`
	ProductCount        int          `json:"ProductCount"`
	IsSortEnabled       bool         `json:"IsSortEnabled"`
	IsPaginationEnabled bool         `json:"IsPaginationEnabled"`
	UrlFriendlyName     string       `json:"UrlFriendlyName"`
	IsSpecial           bool         `json:"IsSpecial"`
	RichRelevanceID     *string      `json:"RichRelevanceId"`
	IsBundle            bool         `json:"IsBundle"`
}

type DepartmentCategoriesList struct {
	Categories []departmentInfo `json:"Categories"`
}

type productID string
type departmentID string

type categoryData []byte
type fruitVegPage []byte

// Prefix for product IDs when exported outside of woolworths-world
const WOOLWORTHS_ID_PREFIX = "woolworths_sku_"

type woolworthsProductInfo struct {
	ID                    productID
	departmentID          departmentID
	departmentDescription string
	Info                  productInfo
	RawJSON               []byte
	Updated               time.Time
}

type categoryRequestBody struct {
	CategoryID                      departmentID `json:"categoryId"`
	PageNumber                      int          `json:"pageNumber"`
	PageSize                        int          `json:"pageSize"`
	SortType                        string       `json:"sortType"`
	URL                             string       `json:"url"`
	Location                        string       `json:"location"`
	FormatObject                    string       `json:"formatObject"`
	IsSpecial                       bool         `json:"isSpecial"`
	IsBundle                        bool         `json:"isBundle"`
	IsMobile                        bool         `json:"isMobile"`
	Filters                         []string     `json:"filters"`
	Token                           string       `json:"token"`
	GPBoost                         int          `json:"gpBoost"`
	IsHideUnavailableProducts       bool         `json:"isHideUnavailableProducts"`
	IsRegisteredRewardCardPromotion bool         `json:"isRegisteredRewardCardPromotion"`
	EnableAdReRanking               bool         `json:"enableAdReRanking"`
	GroupEdmVariants                bool         `json:"groupEdmVariants"`
	CategoryVersion                 string       `json:"categoryVersion"`
}

type productInfo struct {
	Context                   string      `json:"@context"`
	Type                      string      `json:"@type"`
	ID                        interface{} `json:"@id"`
	Name                      string      `json:"name"`
	Description               string      `json:"description"`
	AdditionalType            interface{} `json:"additionalType"`
	AlternateName             interface{} `json:"alternateName"`
	DisambiguatingDescription interface{} `json:"disambiguatingDescription"`
	Identifier                interface{} `json:"identifier"`
	Image                     string      `json:"image"`
	MainEntityOfPage          interface{} `json:"mainEntityOfPage"`
	PotentialAction           interface{} `json:"potentialAction"`
	SameAs                    interface{} `json:"sameAs"`
	URL                       interface{} `json:"url"`
	AdditionalProperty        interface{} `json:"additionalProperty"`
	AggregateRating           interface{} `json:"aggregateRating"`
	Audience                  interface{} `json:"audience"`
	Award                     interface{} `json:"award"`
	Brand                     brand       `json:"brand"`
	Category                  interface{} `json:"category"`
	Color                     interface{} `json:"color"`
	Depth                     interface{} `json:"depth"`
	Gtin12                    interface{} `json:"gtin12"`
	Gtin13                    string      `json:"gtin13"`
	Gtin14                    interface{} `json:"gtin14"`
	Gtin8                     interface{} `json:"gtin8"`
	Height                    interface{} `json:"height"`
	IsAccessoryOrSparePartFor interface{} `json:"isAccessoryOrSparePartFor"`
	IsConsumableFor           interface{} `json:"isConsumableFor"`
	IsRelatedTo               interface{} `json:"isRelatedTo"`
	IsSimilarTo               interface{} `json:"isSimilarTo"`
	ItemCondition             interface{} `json:"itemCondition"`
	Logo                      interface{} `json:"logo"`
	Manufacturer              interface{} `json:"manufacturer"`
	Material                  interface{} `json:"material"`
	Model                     interface{} `json:"model"`
	Mpn                       interface{} `json:"mpn"`
	Offers                    offer       `json:"offers"`
	ProductID                 interface{} `json:"productID"`
	ProductionDate            interface{} `json:"productionDate"`
	PurchaseDate              interface{} `json:"purchaseDate"`
	ReleaseDate               interface{} `json:"releaseDate"`
	Review                    interface{} `json:"review"`
	Sku                       string      `json:"sku"`
	Weight                    float32     `json:"weight"`
	Width                     interface{} `json:"width"`
}

type brand struct {
	Context                   string      `json:"@context"`
	Type                      string      `json:"@type"`
	ID                        interface{} `json:"@id"`
	Name                      string      `json:"name"`
	Description               interface{} `json:"description"`
	AdditionalType            interface{} `json:"additionalType"`
	AlternateName             interface{} `json:"alternateName"`
	DisambiguatingDescription interface{} `json:"disambiguatingDescription"`
	Identifier                interface{} `json:"identifier"`
	Image                     interface{} `json:"image"`
	MainEntityOfPage          interface{} `json:"mainEntityOfPage"`
	PotentialAction           interface{} `json:"potentialAction"`
	SameAs                    interface{} `json:"sameAs"`
	URL                       interface{} `json:"url"`
	ActionableFeedbackPolicy  interface{} `json:"actionableFeedbackPolicy"`
	Address                   interface{} `json:"address"`
	AggregateRating           interface{} `json:"aggregateRating"`
	Alumni                    interface{} `json:"alumni"`
	AreaServed                interface{} `json:"areaServed"`
	Award                     interface{} `json:"award"`
	Brand                     interface{} `json:"brand"`
	ContactPoint              interface{} `json:"contactPoint"`
	CorrectionsPolicy         interface{} `json:"correctionsPolicy"`
	Department                interface{} `json:"department"`
	DissolutionDate           interface{} `json:"dissolutionDate"`
	DiversityPolicy           interface{} `json:"diversityPolicy"`
	Duns                      interface{} `json:"duns"`
	Email                     interface{} `json:"email"`
	Employee                  interface{} `json:"employee"`
	EthicsPolicy              interface{} `json:"ethicsPolicy"`
	Event                     interface{} `json:"event"`
	FaxNumber                 interface{} `json:"faxNumber"`
	Founder                   interface{} `json:"founder"`
	FoundingDate              interface{} `json:"foundingDate"`
	FoundingLocation          interface{} `json:"foundingLocation"`
	Funder                    interface{} `json:"funder"`
	GlobalLocationNumber      interface{} `json:"globalLocationNumber"`
	HasOfferCatalog           interface{} `json:"hasOfferCatalog"`
	HasPOS                    interface{} `json:"hasPOS"`
	IsicV4                    interface{} `json:"isicV4"`
	LegalName                 interface{} `json:"legalName"`
	LeiCode                   interface{} `json:"leiCode"`
	Location                  interface{} `json:"location"`
	Logo                      interface{} `json:"logo"`
	MakesOffer                interface{} `json:"makesOffer"`
	Member                    interface{} `json:"member"`
	MemberOf                  interface{} `json:"memberOf"`
	Naics                     interface{} `json:"naics"`
	NumberOfEmployees         interface{} `json:"numberOfEmployees"`
	Owns                      interface{} `json:"owns"`
	ParentOrganization        interface{} `json:"parentOrganization"`
	PublishingPrinciples      interface{} `json:"publishingPrinciples"`
	Review                    interface{} `json:"review"`
	Seeks                     interface{} `json:"seeks"`
	Sponsor                   interface{} `json:"sponsor"`
	SubOrganization           interface{} `json:"subOrganization"`
	TaxID                     interface{} `json:"taxID"`
	Telephone                 interface{} `json:"telephone"`
	UnnamedSourcesPolicy      interface{} `json:"unnamedSourcesPolicy"`
	VatID                     interface{} `json:"vatID"`
}

type offer struct {
	Context                   string          `json:"@context"`
	Type                      string          `json:"@type"`
	ID                        interface{}     `json:"@id"`
	Name                      interface{}     `json:"name"`
	Description               interface{}     `json:"description"`
	AdditionalType            interface{}     `json:"additionalType"`
	AlternateName             interface{}     `json:"alternateName"`
	DisambiguatingDescription interface{}     `json:"disambiguatingDescription"`
	Identifier                interface{}     `json:"identifier"`
	Image                     interface{}     `json:"image"`
	MainEntityOfPage          interface{}     `json:"mainEntityOfPage"`
	PotentialAction           *buyAction      `json:"potentialAction"`
	SameAs                    interface{}     `json:"sameAs"`
	URL                       interface{}     `json:"url"`
	AcceptedPaymentMethod     interface{}     `json:"acceptedPaymentMethod"`
	AddOn                     interface{}     `json:"addOn"`
	AdvanceBookingRequirement interface{}     `json:"advanceBookingRequirement"`
	AggregateRating           interface{}     `json:"aggregateRating"`
	AreaServed                interface{}     `json:"areaServed"`
	Availability              string          `json:"availability"`
	AvailabilityEnds          interface{}     `json:"availabilityEnds"`
	AvailabilityStarts        interface{}     `json:"availabilityStarts"`
	AvailableAtOrFrom         interface{}     `json:"availableAtOrFrom"`
	AvailableDeliveryMethod   interface{}     `json:"availableDeliveryMethod"`
	BusinessFunction          interface{}     `json:"businessFunction"`
	Category                  interface{}     `json:"category"`
	DeliveryLeadTime          interface{}     `json:"deliveryLeadTime"`
	EligibleCustomerType      interface{}     `json:"eligibleCustomerType"`
	EligibleDuration          interface{}     `json:"eligibleDuration"`
	EligibleQuantity          interface{}     `json:"eligibleQuantity"`
	EligibleRegion            interface{}     `json:"eligibleRegion"`
	EligibleTransactionVolume interface{}     `json:"eligibleTransactionVolume"`
	Gtin12                    interface{}     `json:"gtin12"`
	Gtin13                    interface{}     `json:"gtin13"`
	Gtin14                    interface{}     `json:"gtin14"`
	Gtin8                     interface{}     `json:"gtin8"`
	IncludesObject            interface{}     `json:"includesObject"`
	IneligibleRegion          interface{}     `json:"ineligibleRegion"`
	InventoryLevel            interface{}     `json:"inventoryLevel"`
	ItemCondition             string          `json:"itemCondition"`
	ItemOffered               interface{}     `json:"itemOffered"`
	Mpn                       interface{}     `json:"mpn"`
	OfferedBy                 interface{}     `json:"offeredBy"`
	Price                     decimal.Decimal `json:"price"`
	PriceCurrency             string          `json:"priceCurrency"`
	PriceSpecification        interface{}     `json:"priceSpecification"`
	PriceValidUntil           interface{}     `json:"priceValidUntil"`
	Review                    interface{}     `json:"review"`
	Seller                    interface{}     `json:"seller"`
	SerialNumber              interface{}     `json:"serialNumber"`
	Sku                       interface{}     `json:"sku"`
	ValidFrom                 interface{}     `json:"validFrom"`
	ValidThrough              interface{}     `json:"validThrough"`
	Warranty                  interface{}     `json:"warranty"`
}

type buyAction struct {
	Context                   string      `json:"@context"`
	Type                      string      `json:"@type"`
	ID                        interface{} `json:"@id"`
	Name                      interface{} `json:"name"`
	Description               interface{} `json:"description"`
	AdditionalType            interface{} `json:"additionalType"`
	AlternateName             interface{} `json:"alternateName"`
	DisambiguatingDescription interface{} `json:"disambiguatingDescription"`
	Identifier                interface{} `json:"identifier"`
	Image                     interface{} `json:"image"`
	MainEntityOfPage          interface{} `json:"mainEntityOfPage"`
	PotentialAction           interface{} `json:"potentialAction"`
	SameAs                    interface{} `json:"sameAs"`
	URL                       interface{} `json:"url"`
	ActionStatus              interface{} `json:"actionStatus"`
	Agent                     interface{} `json:"agent"`
	EndTime                   interface{} `json:"endTime"`
	Error                     interface{} `json:"error"`
	Instrument                interface{} `json:"instrument"`
	Location                  interface{} `json:"location"`
	Object                    interface{} `json:"object"`
	Participant               interface{} `json:"participant"`
	Result                    interface{} `json:"result"`
	StartTime                 interface{} `json:"startTime"`
	Target                    interface{} `json:"target"`
}
