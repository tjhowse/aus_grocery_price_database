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
	Updated             time.Time    // Excluded from JSON deserialisation
}

type DepartmentCategoriesList struct {
	Categories []departmentInfo `json:"Categories"`
}

type productID string
type departmentID string

type departmentPage struct {
	ID   departmentID
	page int
}

type categoryData []byte
type fruitVegPage []byte

// Prefix for product IDs when exported outside of woolworths-world
const WOOLWORTHS_ID_PREFIX = "woolworths_sku_"
const PRODUCTS_PER_PAGE = 36

type woolworthsProductInfo struct {
	ID                    productID
	departmentID          departmentID
	departmentDescription string
	Info                  productListPageProduct
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

type productListPageProduct struct {
	TileID                    int             `json:"TileID"`
	Stockcode                 int             `json:"Stockcode"`
	Barcode                   string          `json:"Barcode"`
	GtinFormat                int             `json:"GtinFormat"`
	CupPrice                  float64         `json:"CupPrice"`
	InstoreCupPrice           float64         `json:"InstoreCupPrice"`
	CupMeasure                string          `json:"CupMeasure"`
	CupString                 string          `json:"CupString"`
	InstoreCupString          string          `json:"InstoreCupString"`
	HasCupPrice               bool            `json:"HasCupPrice"`
	InstoreHasCupPrice        bool            `json:"InstoreHasCupPrice"`
	Price                     decimal.Decimal `json:"Price"`
	InstorePrice              float64         `json:"InstorePrice"`
	Name                      string          `json:"Name"`
	DisplayName               string          `json:"DisplayName"`
	URLFriendlyName           string          `json:"UrlFriendlyName"`
	Description               string          `json:"Description"`
	SmallImageFile            string          `json:"SmallImageFile"`
	MediumImageFile           string          `json:"MediumImageFile"`
	LargeImageFile            string          `json:"LargeImageFile"`
	IsNew                     bool            `json:"IsNew"`
	IsHalfPrice               bool            `json:"IsHalfPrice"`
	IsOnlineOnly              bool            `json:"IsOnlineOnly"`
	IsOnSpecial               bool            `json:"IsOnSpecial"`
	InstoreIsOnSpecial        bool            `json:"InstoreIsOnSpecial"`
	IsEdrSpecial              bool            `json:"IsEdrSpecial"`
	SavingsAmount             float64         `json:"SavingsAmount"`
	InstoreSavingsAmount      float64         `json:"InstoreSavingsAmount"`
	WasPrice                  float64         `json:"WasPrice"`
	InstoreWasPrice           float64         `json:"InstoreWasPrice"`
	QuantityInTrolley         int             `json:"QuantityInTrolley"`
	Unit                      string          `json:"Unit"`
	MinimumQuantity           float32         `json:"MinimumQuantity"`
	HasBeenBoughtBefore       bool            `json:"HasBeenBoughtBefore"`
	IsInTrolley               bool            `json:"IsInTrolley"`
	Source                    string          `json:"Source"`
	SupplyLimit               float32         `json:"SupplyLimit"`
	ProductLimit              int             `json:"ProductLimit"`
	MaxSupplyLimitMessage     string          `json:"MaxSupplyLimitMessage"`
	IsRanged                  bool            `json:"IsRanged"`
	IsInStock                 bool            `json:"IsInStock"`
	PackageSize               string          `json:"PackageSize"`
	IsPmDelivery              bool            `json:"IsPmDelivery"`
	IsForCollection           bool            `json:"IsForCollection"`
	IsForDelivery             bool            `json:"IsForDelivery"`
	IsForExpress              bool            `json:"IsForExpress"`
	ProductRestrictionMessage interface{}     `json:"ProductRestrictionMessage"`
	ProductWarningMessage     interface{}     `json:"ProductWarningMessage"`
	CentreTag                 struct {
		TagContent                      interface{} `json:"TagContent"`
		TagLink                         interface{} `json:"TagLink"`
		FallbackText                    interface{} `json:"FallbackText"`
		TagType                         string      `json:"TagType"`
		MultibuyData                    interface{} `json:"MultibuyData"`
		MemberPriceData                 interface{} `json:"MemberPriceData"`
		TagContentText                  interface{} `json:"TagContentText"`
		DualImageTagContent             interface{} `json:"DualImageTagContent"`
		PromotionType                   string      `json:"PromotionType"`
		IsRegisteredRewardCardPromotion bool        `json:"IsRegisteredRewardCardPromotion"`
	} `json:"CentreTag"`
	IsCentreTag bool `json:"IsCentreTag"`
	ImageTag    struct {
		TagContent                      string      `json:"TagContent"`
		TagLink                         interface{} `json:"TagLink"`
		FallbackText                    string      `json:"FallbackText"`
		TagType                         string      `json:"TagType"`
		MultibuyData                    interface{} `json:"MultibuyData"`
		MemberPriceData                 interface{} `json:"MemberPriceData"`
		TagContentText                  interface{} `json:"TagContentText"`
		DualImageTagContent             interface{} `json:"DualImageTagContent"`
		PromotionType                   string      `json:"PromotionType"`
		IsRegisteredRewardCardPromotion bool        `json:"IsRegisteredRewardCardPromotion"`
	} `json:"ImageTag"`
	HeaderTag                    interface{} `json:"HeaderTag"`
	HasHeaderTag                 bool        `json:"HasHeaderTag"`
	UnitWeightInGrams            int         `json:"UnitWeightInGrams"`
	SupplyLimitMessage           string      `json:"SupplyLimitMessage"`
	SmallFormatDescription       string      `json:"SmallFormatDescription"`
	FullDescription              string      `json:"FullDescription"`
	IsAvailable                  bool        `json:"IsAvailable"`
	InstoreIsAvailable           bool        `json:"InstoreIsAvailable"`
	IsPurchasable                bool        `json:"IsPurchasable"`
	InstoreIsPurchasable         bool        `json:"InstoreIsPurchasable"`
	AgeRestricted                bool        `json:"AgeRestricted"`
	DisplayQuantity              float32     `json:"DisplayQuantity"`
	RichDescription              interface{} `json:"RichDescription"`
	HideWasSavedPrice            bool        `json:"HideWasSavedPrice"`
	SapCategories                interface{} `json:"SapCategories"`
	Brand                        interface{} `json:"Brand"`
	IsRestrictedByDeliveryMethod bool        `json:"IsRestrictedByDeliveryMethod"`
	FooterTag                    struct {
		TagContent                      interface{} `json:"TagContent"`
		TagLink                         interface{} `json:"TagLink"`
		FallbackText                    interface{} `json:"FallbackText"`
		TagType                         string      `json:"TagType"`
		MultibuyData                    interface{} `json:"MultibuyData"`
		MemberPriceData                 interface{} `json:"MemberPriceData"`
		TagContentText                  interface{} `json:"TagContentText"`
		DualImageTagContent             interface{} `json:"DualImageTagContent"`
		PromotionType                   string      `json:"PromotionType"`
		IsRegisteredRewardCardPromotion bool        `json:"IsRegisteredRewardCardPromotion"`
	} `json:"FooterTag"`
	IsFooterEnabled      bool        `json:"IsFooterEnabled"`
	Diagnostics          string      `json:"Diagnostics"`
	IsBundle             bool        `json:"IsBundle"`
	IsInFamily           bool        `json:"IsInFamily"`
	ChildProducts        interface{} `json:"ChildProducts"`
	URLOverride          interface{} `json:"UrlOverride"`
	AdditionalAttributes struct {
		Boxedcontents                interface{} `json:"boxedcontents"`
		Addedvitaminsandminerals     string      `json:"addedvitaminsandminerals"`
		Sapdepartmentname            string      `json:"sapdepartmentname"`
		Spf                          interface{} `json:"spf"`
		Haircolour                   interface{} `json:"haircolour"`
		Lifestyleanddietarystatement interface{} `json:"lifestyleanddietarystatement"`
		Sapcategoryname              string      `json:"sapcategoryname"`
		Skintype                     interface{} `json:"skintype"`
		Importantinformation         interface{} `json:"importantinformation"`
		Allergystatement             interface{} `json:"allergystatement"`
		Productdepthmm               interface{} `json:"productdepthmm"`
		Skincondition                interface{} `json:"skincondition"`
		Ophthalmologistapproved      interface{} `json:"ophthalmologistapproved"`
		Healthstarrating             string      `json:"healthstarrating"`
		Hairtype                     interface{} `json:"hairtype"`
		FragranceFree                interface{} `json:"fragrance-free"`
		Sapsegmentname               string      `json:"sapsegmentname"`
		Suitablefor                  interface{} `json:"suitablefor"`
		PiesProductDepartmentsjson   string      `json:"PiesProductDepartmentsjson"`
		Piessubcategorynamesjson     string      `json:"piessubcategorynamesjson"`
		Sapsegmentno                 string      `json:"sapsegmentno"`
		Productwidthmm               interface{} `json:"productwidthmm"`
		Contains                     interface{} `json:"contains"`
		Sapsubcategoryname           string      `json:"sapsubcategoryname"`
		Dermatologisttested          interface{} `json:"dermatologisttested"`
		WoolProductpackaging         interface{} `json:"wool_productpackaging"`
		Dermatologicallyapproved     interface{} `json:"dermatologicallyapproved"`
		Specialsgroupid              interface{} `json:"specialsgroupid"`
		Productimages                string      `json:"productimages"`
		Productheightmm              interface{} `json:"productheightmm"`
		RRHidereviews                interface{} `json:"r&r_hidereviews"`
		Microwavesafe                string      `json:"microwavesafe"`
		PabaFree                     interface{} `json:"paba-free"`
		Lifestyleclaim               interface{} `json:"lifestyleclaim"`
		Alcoholfree                  interface{} `json:"alcoholfree"`
		Tgawarning                   interface{} `json:"tgawarning"`
		Activeconstituents           interface{} `json:"activeconstituents"`
		Microwaveable                string      `json:"microwaveable"`
		SoapFree                     interface{} `json:"soap-free"`
		Countryoforigin              interface{} `json:"countryoforigin"`
		Isexcludedfromsubstitution   string      `json:"isexcludedfromsubstitution"`
		Productimagecount            string      `json:"productimagecount"`
		RRLoggedinreviews            interface{} `json:"r&r_loggedinreviews"`
		AntiDandruff                 interface{} `json:"anti-dandruff"`
		ServingsizeTotalNip          interface{} `json:"servingsize-total-nip"`
		Tgahealthwarninglink         interface{} `json:"tgahealthwarninglink"`
		Allergenmaybepresent         interface{} `json:"allergenmaybepresent"`
		PiesProductDepartmentNodeID  string      `json:"PiesProductDepartmentNodeId"`
		Parabenfree                  string      `json:"parabenfree"`
		Vendorarticleid              interface{} `json:"vendorarticleid"`
		Containsgluten               string      `json:"containsgluten"`
		Containsnuts                 string      `json:"containsnuts"`
		Ingredients                  interface{} `json:"ingredients"`
		Colour                       interface{} `json:"colour"`
		Manufacturer                 interface{} `json:"manufacturer"`
		Sapcategoryno                string      `json:"sapcategoryno"`
		Storageinstructions          interface{} `json:"storageinstructions"`
		Tgawarnings                  interface{} `json:"tgawarnings"`
		Piesdepartmentnamesjson      string      `json:"piesdepartmentnamesjson"`
		Brand                        interface{} `json:"brand"`
		Oilfree                      interface{} `json:"oilfree"`
		Fragrance                    interface{} `json:"fragrance"`
		Antibacterial                string      `json:"antibacterial"`
		NonComedogenic               interface{} `json:"non-comedogenic"`
		Antiseptic                   string      `json:"antiseptic"`
		Bpafree                      string      `json:"bpafree"`
		Vendorcostprice              interface{} `json:"vendorcostprice"`
		Description                  string      `json:"description"`
		Sweatresistant               interface{} `json:"sweatresistant"`
		Sapsubcategoryno             string      `json:"sapsubcategoryno"`
		Antioxidant                  string      `json:"antioxidant"`
		Claims                       interface{} `json:"claims"`
		Phbalanced                   interface{} `json:"phbalanced"`
		WoolDietaryclaim             interface{} `json:"wool_dietaryclaim"`
		Ophthalmologisttested        interface{} `json:"ophthalmologisttested"`
		Sulfatefree                  string      `json:"sulfatefree"`
		Piescategorynamesjson        string      `json:"piescategorynamesjson"`
		ServingsperpackTotalNip      interface{} `json:"servingsperpack-total-nip"`
		Nutritionalinformation       interface{} `json:"nutritionalinformation"`
		Ovencook                     string      `json:"ovencook"`
		Vegetarian                   string      `json:"vegetarian"`
		HypoAllergenic               interface{} `json:"hypo-allergenic"`
		Timer                        interface{} `json:"timer"`
		Dermatologistrecommended     interface{} `json:"dermatologistrecommended"`
		Sapdepartmentno              string      `json:"sapdepartmentno"`
		Allergencontains             interface{} `json:"allergencontains"`
		Waterresistant               interface{} `json:"waterresistant"`
		Friendlydisclaimer           interface{} `json:"friendlydisclaimer"`
		Recyclableinformation        interface{} `json:"recyclableinformation"`
		Usageinstructions            interface{} `json:"usageinstructions"`
		Freezable                    string      `json:"freezable"`
	} `json:"AdditionalAttributes"`
	DetailsImagePaths []string    `json:"DetailsImagePaths"`
	Variety           interface{} `json:"Variety"`
	Rating            struct {
		ReviewCount         int `json:"ReviewCount"`
		RatingCount         int `json:"RatingCount"`
		RatingSum           int `json:"RatingSum"`
		OneStarCount        int `json:"OneStarCount"`
		TwoStarCount        int `json:"TwoStarCount"`
		ThreeStarCount      int `json:"ThreeStarCount"`
		FourStarCount       int `json:"FourStarCount"`
		FiveStarCount       int `json:"FiveStarCount"`
		Average             int `json:"Average"`
		OneStarPercentage   int `json:"OneStarPercentage"`
		TwoStarPercentage   int `json:"TwoStarPercentage"`
		ThreeStarPercentage int `json:"ThreeStarPercentage"`
		FourStarPercentage  int `json:"FourStarPercentage"`
		FiveStarPercentage  int `json:"FiveStarPercentage"`
	} `json:"Rating"`
	HasProductSubs        bool        `json:"HasProductSubs"`
	IsSponsoredAd         bool        `json:"IsSponsoredAd"`
	AdID                  interface{} `json:"AdID"`
	AdIndex               interface{} `json:"AdIndex"`
	AdStatus              interface{} `json:"AdStatus"`
	IsMarketProduct       bool        `json:"IsMarketProduct"`
	IsGiftable            bool        `json:"IsGiftable"`
	Vendor                interface{} `json:"Vendor"`
	Untraceable           bool        `json:"Untraceable"`
	ThirdPartyProductInfo interface{} `json:"ThirdPartyProductInfo"`
	MarketFeatures        interface{} `json:"MarketFeatures"`
	MarketSpecifications  interface{} `json:"MarketSpecifications"`
	SupplyLimitSource     string      `json:"SupplyLimitSource"`
	Tags                  []struct {
		Content struct {
			Type       string `json:"Type"`
			Position   string `json:"Position"`
			Attributes struct {
				ImagePath    string `json:"ImagePath"`
				FallbackText string `json:"FallbackText"`
			} `json:"Attributes"`
		} `json:"Content"`
		TemplateID interface{} `json:"TemplateId"`
		Metadata   interface{} `json:"Metadata"`
	} `json:"Tags"`
	IsPersonalisedByPurchaseHistory bool        `json:"IsPersonalisedByPurchaseHistory"`
	IsFromFacetedSearch             bool        `json:"IsFromFacetedSearch"`
	NextAvailabilityDate            time.Time   `json:"NextAvailabilityDate"`
	NumberOfSubstitutes             int         `json:"NumberOfSubstitutes"`
	IsPrimaryVariant                bool        `json:"IsPrimaryVariant"`
	VariantGroupID                  int         `json:"VariantGroupId"`
	HasVariants                     bool        `json:"HasVariants"`
	VariantTitle                    interface{} `json:"VariantTitle"`
	IsTobacco                       bool        `json:"IsTobacco"`
	IsB2BExtendedRangeSapCategory   bool        `json:"IsB2BExtendedRangeSapCategory"`
}

type productListPageProducts struct {
	Products    []productListPageProduct `json:"Products"`
	Name        string                   `json:"Name"`
	DisplayName string                   `json:"DisplayName"`
}

type productListPage struct {
	SeoMetaTags struct {
		Title           string        `json:"Title"`
		MetaDescription string        `json:"MetaDescription"`
		Groups          []interface{} `json:"Groups"`
	} `json:"SeoMetaTags"`
	Bundles                []productListPageProducts `json:"Bundles"`
	TotalRecordCount       int                       `json:"TotalRecordCount"`
	UpperDynamicContent    interface{}               `json:"UpperDynamicContent"`
	LowerDynamicContent    interface{}               `json:"LowerDynamicContent"`
	RichRelevancePlacement struct {
		PlacementName         interface{}   `json:"placement_name"`
		Message               interface{}   `json:"message"`
		Products              []interface{} `json:"Products"`
		Items                 []interface{} `json:"Items"`
		StockcodesForDiscover []interface{} `json:"StockcodesForDiscover"`
	} `json:"RichRelevancePlacement"`
	Aggregations []struct {
		Name           string `json:"Name"`
		DisplayName    string `json:"DisplayName"`
		Type           string `json:"Type"`
		FilterType     string `json:"FilterType"`
		FilterDataType string `json:"FilterDataType"`
		Results        []struct {
			Name              string `json:"Name"`
			Term              string `json:"Term"`
			ExtraOutputFields struct {
			} `json:"ExtraOutputFields"`
			Min               interface{} `json:"Min"`
			Max               interface{} `json:"Max"`
			Applied           bool        `json:"Applied"`
			Count             int         `json:"Count"`
			Statement         string      `json:"Statement"`
			DisplayCoachMarks bool        `json:"DisplayCoachMarks"`
		} `json:"Results"`
		ResultsGrouped    interface{} `json:"ResultsGrouped"`
		State             string      `json:"State"`
		Rank              int         `json:"Rank"`
		AdditionalResults bool        `json:"AdditionalResults"`
		DesignType        string      `json:"DesignType"`
		ShowFilter        bool        `json:"ShowFilter"`
		Statement         string      `json:"Statement"`
		DisplayCoachMarks bool        `json:"DisplayCoachMarks"`
		DisplayIcons      bool        `json:"DisplayIcons"`
	} `json:"Aggregations"`
	HasRewardsCard  bool `json:"HasRewardsCard"`
	HasTobaccoItems bool `json:"HasTobaccoItems"`
	Success         bool `json:"Success"`
}
