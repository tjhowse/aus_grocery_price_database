package coles

import (
	"time"

	"github.com/shopspring/decimal"
)

type productID string
type departmentID string

const COLES_ID_PREFIX = "coles_id_"

type colesProductInfo struct {
	ID                    productID
	departmentID          string
	departmentDescription string
	Info                  productListPageProduct
	WeightGrams           int
	PreviousPrice         decimal.Decimal
	RawJSON               []byte
	Updated               time.Time
}

type productListPageProductPricing struct {
	Now  decimal.Decimal `json:"now"`
	Was  decimal.Decimal `json:"was"`
	Unit struct {
		Quantity          float64         `json:"quantity"`
		OfMeasureQuantity float64         `json:"ofMeasureQuantity"`
		OfMeasureUnits    string          `json:"ofMeasureUnits"`
		Price             decimal.Decimal `json:"price"`
		OfMeasureType     string          `json:"ofMeasureType"`
		IsWeighted        bool            `json:"isWeighted"`
	} `json:"unit"`
	Comparable    string `json:"comparable"`
	PromotionType string `json:"promotionType"`
	OnlineSpecial bool   `json:"onlineSpecial"`
}

type productListPageProduct struct {
	Type             string      `json:"_type"`
	ID               int         `json:"id,omitempty"`
	AdID             interface{} `json:"adId"`
	AdSource         interface{} `json:"adSource"`
	Featured         bool        `json:"featured,omitempty"`
	Name             string      `json:"name,omitempty"`
	Brand            string      `json:"brand,omitempty"`
	Description      string      `json:"description,omitempty"`
	Size             string      `json:"size,omitempty"`
	Availability     bool        `json:"availability,omitempty"`
	AvailabilityType string      `json:"availabilityType,omitempty"`
	ImageUris        []struct {
		AltText string `json:"altText"`
		Type    string `json:"type"`
		URI     string `json:"uri"`
	} `json:"imageUris,omitempty"`
	Locations []struct {
		AisleSide   interface{} `json:"aisleSide"`
		Description string      `json:"description"`
		Facing      int         `json:"facing"`
		Aisle       interface{} `json:"aisle"`
		Order       int         `json:"order"`
		Shelf       interface{} `json:"shelf"`
	} `json:"locations,omitempty"`
	Restrictions struct {
		RetailLimit               int      `json:"retailLimit"`
		PromotionalLimit          int      `json:"promotionalLimit"`
		LiquorAgeRestrictionFlag  bool     `json:"liquorAgeRestrictionFlag"`
		TobaccoAgeRestrictionFlag bool     `json:"tobaccoAgeRestrictionFlag"`
		RestrictedByOrganisation  bool     `json:"restrictedByOrganisation"`
		Delivery                  []string `json:"delivery"`
	} `json:"restrictions,omitempty"`
	MerchandiseHeir struct {
		TradeProfitCentre string `json:"tradeProfitCentre"`
		CategoryGroup     string `json:"categoryGroup"`
		Category          string `json:"category"`
		SubCategory       string `json:"subCategory"`
		ClassName         string `json:"className"`
	} `json:"merchandiseHeir,omitempty"`
	OnlineHeirs []struct {
		Aisle         string `json:"aisle"`
		Category      string `json:"category"`
		SubCategory   string `json:"subCategory"`
		CategoryID    string `json:"categoryId"`
		AisleID       string `json:"aisleId"`
		SubCategoryID string `json:"subCategoryId"`
	} `json:"onlineHeirs,omitempty"`
	Pricing                          productListPageProductPricing `json:"pricing,omitempty"`
	CampaignName                     string                        `json:"campaignName,omitempty"`
	Expiry                           interface{}                   `json:"expiry,omitempty"`
	HeadingText                      interface{}                   `json:"headingText,omitempty"`
	BannerText                       string                        `json:"bannerText,omitempty"`
	BannerTextColour                 interface{}                   `json:"bannerTextColour,omitempty"`
	CtaFlag                          interface{}                   `json:"ctaFlag,omitempty"`
	CtaText                          string                        `json:"ctaText,omitempty"`
	CtaTextAccessibility             string                        `json:"ctaTextAccessibility,omitempty"`
	CtaLink                          string                        `json:"ctaLink,omitempty"`
	BackgroundColour                 interface{}                   `json:"backgroundColour,omitempty"`
	BackgroundImage                  string                        `json:"backgroundImage,omitempty"`
	BackgroundImagePosition          interface{}                   `json:"backgroundImagePosition,omitempty"`
	SecondaryBackgroundImage         interface{}                   `json:"secondaryBackgroundImage,omitempty"`
	SecondaryBackgroundImagePosition interface{}                   `json:"secondaryBackgroundImagePosition,omitempty"`
	HeroImage                        interface{}                   `json:"heroImage,omitempty"`
	HeroImageAltText                 interface{}                   `json:"heroImageAltText,omitempty"`
	SecondaryHeroImage               string                        `json:"secondaryHeroImage,omitempty"`
	SecondaryHeroImageAltText        string                        `json:"secondaryHeroImageAltText,omitempty"`
	ProductIds                       []string                      `json:"productIds,omitempty"`
	AdditionalFields                 []struct {
		ID    string `json:"id"`
		Value string `json:"value"`
	} `json:"additionalFields,omitempty"`
	Continuity struct {
		ContinuityPromotionID   interface{} `json:"continuityPromotionId"`
		CreditsToRedeem         interface{} `json:"creditsToRedeem"`
		BonusAvailable          bool        `json:"bonusAvailable"`
		BonusTimes              int         `json:"bonusTimes"`
		BonusPromoName          string      `json:"bonusPromoName"`
		BonusRoundelDisplayable bool        `json:"bonusRoundelDisplayable"`
		BonusRoundelDescription string      `json:"bonusRoundelDescription"`
	} `json:"continuity,omitempty"`
	TargetProductID int `json:"targetProductId,omitempty"`
	Variations      struct {
		Total int `json:"total"`
	} `json:"variations,omitempty"`
}

type categoryPage struct {
	PageProps struct {
		AssetsURL       string `json:"assetsUrl"`
		SentryTraceData string `json:"_sentryTraceData"`
		SentryBaggage   string `json:"_sentryBaggage"`
		IsMobile        bool   `json:"isMobile"`
		SearchResults   struct {
			DidYouMean      interface{} `json:"didYouMean"`
			NoOfResults     int         `json:"noOfResults"`
			Start           int         `json:"start"`
			PageSize        int         `json:"pageSize"`
			Keyword         interface{} `json:"keyword"`
			ResultType      int         `json:"resultType"`
			AlternateResult bool        `json:"alternateResult"`
			Filters         []struct {
				Name   string `json:"name"`
				Values []struct {
					ID          string `json:"id"`
					DisplayText string `json:"displayText"`
					Count       int    `json:"count"`
					Brand       string `json:"brand,omitempty"`
				} `json:"values"`
			} `json:"filters"`
			Banners []struct {
				Type                      string   `json:"_type"`
				AdID                      string   `json:"adId"`
				AdSource                  string   `json:"adSource"`
				CampaignName              string   `json:"campaignName"`
				BannerText                string   `json:"bannerText"`
				CtaText                   string   `json:"ctaText"`
				CtaTextAccessibility      string   `json:"ctaTextAccessibility"`
				CtaLink                   string   `json:"ctaLink"`
				BackgroundColour          string   `json:"backgroundColour"`
				BackgroundImage           string   `json:"backgroundImage"`
				HeroImage                 string   `json:"heroImage"`
				HeroImageAltText          string   `json:"heroImageAltText"`
				SecondaryHeroImage        string   `json:"secondaryHeroImage"`
				SecondaryHeroImageAltText string   `json:"secondaryHeroImageAltText"`
				ProductIds                []string `json:"productIds"`
				AdditionalFields          []struct {
					ID    string `json:"id"`
					Value string `json:"value"`
				} `json:"additionalFields"`
			} `json:"banners"`
			AdMemoryToken    string `json:"adMemoryToken"`
			PageRestrictions struct {
				TobaccoProducts                  bool `json:"tobaccoProducts"`
				RestrictedByOrganisationProducts bool `json:"restrictedByOrganisationProducts"`
			} `json:"pageRestrictions"`
			Results          []productListPageProduct `json:"results"`
			CatalogGroupView []struct {
				Level            int         `json:"level"`
				Name             string      `json:"name"`
				OriginalName     string      `json:"originalName"`
				SeoToken         string      `json:"seoToken"`
				ID               string      `json:"id"`
				ProductCount     int         `json:"productCount"`
				CatalogGroupView interface{} `json:"catalogGroupView"`
			} `json:"catalogGroupView"`
			ExcludedCatalogGroupView struct {
				ProductCount int `json:"productCount"`
			} `json:"excludedCatalogGroupView"`
		} `json:"searchResults"`
		SearchSessionID    string `json:"searchSessionId"`
		SortByValue        string `json:"sortByValue"`
		IsRestricted       bool   `json:"isRestricted"`
		AemContentFragment struct {
			Data struct {
				CategoryContentDescriptionByPath struct {
					Item struct {
						Description struct {
							HTML string `json:"html"`
						} `json:"description"`
						Tags interface{} `json:"tags"`
					} `json:"item"`
				} `json:"categoryContentDescriptionByPath"`
			} `json:"data"`
		} `json:"aemContentFragment"`
		AemBrandFragment struct {
		} `json:"aemBrandFragment"`
		InitialState struct {
			User struct {
				Error interface{} `json:"error"`
				Auth  struct {
					Authenticated bool `json:"authenticated"`
				} `json:"auth"`
				Account struct {
					Notifications []interface{} `json:"notifications"`
				} `json:"account"`
			} `json:"user"`
			Modal struct {
				Active interface{} `json:"active"`
				State  struct {
				} `json:"state"`
			} `json:"modal"`
			Notifications struct {
				Notifications        []interface{} `json:"notifications"`
				ListNotifications    []interface{} `json:"listNotifications"`
				ShowShoppableWarning bool          `json:"showShoppableWarning"`
			} `json:"notifications"`
			Mpgs struct {
				FormFieldValidity struct {
					CardNumberValidity  string `json:"cardNumberValidity"`
					ExpiryYearValidity  string `json:"expiryYearValidity"`
					ExpiryMonthValidity string `json:"expiryMonthValidity"`
					CvvValidity         string `json:"cvvValidity"`
				} `json:"formFieldValidity"`
				InitStatus      string      `json:"initStatus"`
				SubmitStatus    string      `json:"submitStatus"`
				SuccessData     interface{} `json:"successData"`
				UnexpectedError bool        `json:"unexpectedError"`
				SaveToProfile   bool        `json:"saveToProfile"`
			} `json:"mpgs"`
			Trolley struct {
				Error                                 interface{}   `json:"error"`
				ItemsBeingUpdated                     []interface{} `json:"itemsBeingUpdated"`
				FailedItemGroups                      []interface{} `json:"failedItemGroups"`
				ResolvedProductIdsFromFailedItemGroup []interface{} `json:"resolvedProductIdsFromFailedItemGroup"`
				StoreID                               string        `json:"storeId"`
				Validation                            struct {
					IsValidating     bool        `json:"isValidating"`
					IsValid          bool        `json:"isValid"`
					ValidationErrors interface{} `json:"validationErrors"`
					Error            interface{} `json:"error"`
					RestrictedItems  interface{} `json:"restrictedItems"`
				} `json:"validation"`
				IsSwappingItems                            bool          `json:"isSwappingItems"`
				UpdateQueue                                []interface{} `json:"updateQueue"`
				UpdateQueueCallbacks                       []interface{} `json:"updateQueueCallbacks"`
				ProcessUpdateQueueImmediately              bool          `json:"processUpdateQueueImmediately"`
				IsProcessUpdateQueueErrorNotificationMuted bool          `json:"isProcessUpdateQueueErrorNotificationMuted"`
				FetchContext                               struct {
				} `json:"fetchContext"`
			} `json:"trolley"`
			Drawer struct {
				Active []interface{} `json:"active"`
				State  struct {
				} `json:"state"`
			} `json:"drawer"`
			ShoppingMethod struct {
				IsEditing        bool `json:"isEditing"`
				DidStoreIDChange bool `json:"didStoreIdChange"`
				State            struct {
				} `json:"state"`
			} `json:"shoppingMethod"`
			EnquiryForms struct {
				Ids      []interface{} `json:"ids"`
				Entities struct {
				} `json:"entities"`
			} `json:"enquiryForms"`
			List struct {
				Error               interface{}   `json:"error"`
				PatchListItemsQueue []interface{} `json:"patchListItemsQueue"`
			} `json:"list"`
			Content struct {
				PageCategoryL1           string        `json:"pageCategoryL1"`
				PageCategoryL2           string        `json:"pageCategoryL2"`
				DisplayFilter            bool          `json:"displayFilter"`
				ExpandFilter             []interface{} `json:"expandFilter"`
				NextLevel                bool          `json:"nextLevel"`
				PageTitle                string        `json:"pageTitle"`
				PageType                 string        `json:"pageType"`
				Breadcrumbs              []interface{} `json:"breadcrumbs"`
				RecipeID                 string        `json:"recipeId"`
				IsDisplayShopIngredients bool          `json:"isDisplayShopIngredients"`
				GlobalUrgencyStrip       struct {
				} `json:"globalUrgencyStrip"`
				RecipeServingSize int  `json:"recipeServingSize"`
				DeliveryMethod    bool `json:"deliveryMethod"`
			} `json:"content"`
			SeoJSONLd struct {
				ComponentJSONLd struct {
				} `json:"componentJsonLd"`
				ShowAsJSONLd bool `json:"showAsJsonLd"`
			} `json:"seoJsonLd"`
			BffAPI struct {
				Queries struct {
				} `json:"queries"`
				Mutations struct {
				} `json:"mutations"`
				Provided struct {
				} `json:"provided"`
				Subscriptions struct {
				} `json:"subscriptions"`
				Config struct {
					Online                    bool   `json:"online"`
					Focused                   bool   `json:"focused"`
					MiddlewareRegistered      bool   `json:"middlewareRegistered"`
					RefetchOnFocus            bool   `json:"refetchOnFocus"`
					RefetchOnReconnect        bool   `json:"refetchOnReconnect"`
					RefetchOnMountOrArgChange bool   `json:"refetchOnMountOrArgChange"`
					KeepUnusedDataFor         int    `json:"keepUnusedDataFor"`
					ReducerPath               string `json:"reducerPath"`
				} `json:"config"`
			} `json:"bffApi"`
			AemAPI struct {
				Queries struct {
				} `json:"queries"`
				Mutations struct {
				} `json:"mutations"`
				Provided struct {
				} `json:"provided"`
				Subscriptions struct {
				} `json:"subscriptions"`
				Config struct {
					Online                    bool   `json:"online"`
					Focused                   bool   `json:"focused"`
					MiddlewareRegistered      bool   `json:"middlewareRegistered"`
					RefetchOnFocus            bool   `json:"refetchOnFocus"`
					RefetchOnReconnect        bool   `json:"refetchOnReconnect"`
					RefetchOnMountOrArgChange bool   `json:"refetchOnMountOrArgChange"`
					KeepUnusedDataFor         int    `json:"keepUnusedDataFor"`
					ReducerPath               string `json:"reducerPath"`
				} `json:"config"`
			} `json:"aemApi"`
			EnquiryFormAPI struct {
				Queries struct {
				} `json:"queries"`
				Mutations struct {
				} `json:"mutations"`
				Provided struct {
				} `json:"provided"`
				Subscriptions struct {
				} `json:"subscriptions"`
				Config struct {
					Online                    bool   `json:"online"`
					Focused                   bool   `json:"focused"`
					MiddlewareRegistered      bool   `json:"middlewareRegistered"`
					RefetchOnFocus            bool   `json:"refetchOnFocus"`
					RefetchOnReconnect        bool   `json:"refetchOnReconnect"`
					RefetchOnMountOrArgChange bool   `json:"refetchOnMountOrArgChange"`
					KeepUnusedDataFor         int    `json:"keepUnusedDataFor"`
					ReducerPath               string `json:"reducerPath"`
				} `json:"config"`
			} `json:"enquiryFormApi"`
			B2BFormsAPI struct {
				Queries struct {
				} `json:"queries"`
				Mutations struct {
				} `json:"mutations"`
				Provided struct {
				} `json:"provided"`
				Subscriptions struct {
				} `json:"subscriptions"`
				Config struct {
					Online                    bool   `json:"online"`
					Focused                   bool   `json:"focused"`
					MiddlewareRegistered      bool   `json:"middlewareRegistered"`
					RefetchOnFocus            bool   `json:"refetchOnFocus"`
					RefetchOnReconnect        bool   `json:"refetchOnReconnect"`
					RefetchOnMountOrArgChange bool   `json:"refetchOnMountOrArgChange"`
					KeepUnusedDataFor         int    `json:"keepUnusedDataFor"`
					ReducerPath               string `json:"reducerPath"`
				} `json:"config"`
			} `json:"b2bFormsApi"`
			RadioComplaintsFormAPI struct {
				Queries struct {
				} `json:"queries"`
				Mutations struct {
				} `json:"mutations"`
				Provided struct {
				} `json:"provided"`
				Subscriptions struct {
				} `json:"subscriptions"`
				Config struct {
					Online                    bool   `json:"online"`
					Focused                   bool   `json:"focused"`
					MiddlewareRegistered      bool   `json:"middlewareRegistered"`
					RefetchOnFocus            bool   `json:"refetchOnFocus"`
					RefetchOnReconnect        bool   `json:"refetchOnReconnect"`
					RefetchOnMountOrArgChange bool   `json:"refetchOnMountOrArgChange"`
					KeepUnusedDataFor         int    `json:"keepUnusedDataFor"`
					ReducerPath               string `json:"reducerPath"`
				} `json:"config"`
			} `json:"radioComplaintsFormApi"`
			PsdsFormAPI struct {
				Queries struct {
				} `json:"queries"`
				Mutations struct {
				} `json:"mutations"`
				Provided struct {
				} `json:"provided"`
				Subscriptions struct {
				} `json:"subscriptions"`
				Config struct {
					Online                    bool   `json:"online"`
					Focused                   bool   `json:"focused"`
					MiddlewareRegistered      bool   `json:"middlewareRegistered"`
					RefetchOnFocus            bool   `json:"refetchOnFocus"`
					RefetchOnReconnect        bool   `json:"refetchOnReconnect"`
					RefetchOnMountOrArgChange bool   `json:"refetchOnMountOrArgChange"`
					KeepUnusedDataFor         int    `json:"keepUnusedDataFor"`
					ReducerPath               string `json:"reducerPath"`
				} `json:"config"`
			} `json:"psdsFormApi"`
			AdobeTargetAPI struct {
				Queries struct {
				} `json:"queries"`
				Mutations struct {
				} `json:"mutations"`
				Provided struct {
				} `json:"provided"`
				Subscriptions struct {
				} `json:"subscriptions"`
				Config struct {
					Online                    bool   `json:"online"`
					Focused                   bool   `json:"focused"`
					MiddlewareRegistered      bool   `json:"middlewareRegistered"`
					RefetchOnFocus            bool   `json:"refetchOnFocus"`
					RefetchOnReconnect        bool   `json:"refetchOnReconnect"`
					RefetchOnMountOrArgChange bool   `json:"refetchOnMountOrArgChange"`
					KeepUnusedDataFor         int    `json:"keepUnusedDataFor"`
					ReducerPath               string `json:"reducerPath"`
				} `json:"config"`
			} `json:"adobeTargetApi"`
			AbandonedTrolleyFormAPI struct {
				Queries struct {
				} `json:"queries"`
				Mutations struct {
				} `json:"mutations"`
				Provided struct {
				} `json:"provided"`
				Subscriptions struct {
				} `json:"subscriptions"`
				Config struct {
					Online                    bool   `json:"online"`
					Focused                   bool   `json:"focused"`
					MiddlewareRegistered      bool   `json:"middlewareRegistered"`
					RefetchOnFocus            bool   `json:"refetchOnFocus"`
					RefetchOnReconnect        bool   `json:"refetchOnReconnect"`
					RefetchOnMountOrArgChange bool   `json:"refetchOnMountOrArgChange"`
					KeepUnusedDataFor         int    `json:"keepUnusedDataFor"`
					ReducerPath               string `json:"reducerPath"`
				} `json:"config"`
			} `json:"abandonedTrolleyFormApi"`
			DigitalGraphQLAPI struct {
				Queries struct {
					GetProductCategoriesStoreIDCOL584 struct {
						Status       string `json:"status"`
						EndpointName string `json:"endpointName"`
						RequestID    string `json:"requestId"`
						OriginalArgs struct {
							StoreID string `json:"storeId"`
						} `json:"originalArgs"`
						StartedTimeStamp int64 `json:"startedTimeStamp"`
						Data             struct {
							ProductCategories struct {
								ExcludedCategoryIds []string `json:"excludedCategoryIds"`
								CatalogGroupView    []struct {
									ID               string `json:"id"`
									Level            int    `json:"level"`
									Name             string `json:"name"`
									OriginalName     string `json:"originalName"`
									ProductCount     int    `json:"productCount"`
									SeoToken         string `json:"seoToken"`
									CatalogGroupView []struct {
										ID               string `json:"id"`
										Level            int    `json:"level"`
										Name             string `json:"name"`
										OriginalName     string `json:"originalName"`
										ProductCount     int    `json:"productCount"`
										SeoToken         string `json:"seoToken"`
										CatalogGroupView []struct {
											ID           string `json:"id"`
											Level        int    `json:"level"`
											Name         string `json:"name"`
											OriginalName string `json:"originalName"`
											ProductCount int    `json:"productCount"`
											SeoToken     string `json:"seoToken"`
										} `json:"catalogGroupView"`
									} `json:"catalogGroupView"`
								} `json:"catalogGroupView"`
							} `json:"productCategories"`
						} `json:"data"`
						FulfilledTimeStamp int64 `json:"fulfilledTimeStamp"`
					} `json:"GetProductCategories({"storeId":"COL:584"})"`
				} `json:"queries"`
				Mutations struct {
				} `json:"mutations"`
				Provided struct {
				} `json:"provided"`
				Subscriptions struct {
					GetProductCategoriesStoreIDCOL584 struct {
						KksWf342KsqgvZEM7FB struct {
						} `json:"kksWf342ksqgvZ_-EM7FB"`
					} `json:"GetProductCategories({"storeId":"COL:584"})"`
				} `json:"subscriptions"`
				Config struct {
					Online                    bool   `json:"online"`
					Focused                   bool   `json:"focused"`
					MiddlewareRegistered      bool   `json:"middlewareRegistered"`
					RefetchOnFocus            bool   `json:"refetchOnFocus"`
					RefetchOnReconnect        bool   `json:"refetchOnReconnect"`
					RefetchOnMountOrArgChange bool   `json:"refetchOnMountOrArgChange"`
					KeepUnusedDataFor         int    `json:"keepUnusedDataFor"`
					ReducerPath               string `json:"reducerPath"`
				} `json:"config"`
			} `json:"digitalGraphQLApi"`
			NextAPI struct {
				Queries struct {
				} `json:"queries"`
				Mutations struct {
				} `json:"mutations"`
				Provided struct {
				} `json:"provided"`
				Subscriptions struct {
				} `json:"subscriptions"`
				Config struct {
					Online                    bool   `json:"online"`
					Focused                   bool   `json:"focused"`
					MiddlewareRegistered      bool   `json:"middlewareRegistered"`
					RefetchOnFocus            bool   `json:"refetchOnFocus"`
					RefetchOnReconnect        bool   `json:"refetchOnReconnect"`
					RefetchOnMountOrArgChange bool   `json:"refetchOnMountOrArgChange"`
					KeepUnusedDataFor         int    `json:"keepUnusedDataFor"`
					ReducerPath               string `json:"reducerPath"`
				} `json:"config"`
			} `json:"nextApi"`
		} `json:"initialState"`
	} `json:"pageProps"`
	InitialState struct {
		User struct {
			Error interface{} `json:"error"`
			Auth  struct {
				Authenticated bool `json:"authenticated"`
			} `json:"auth"`
			Account struct {
				Notifications []interface{} `json:"notifications"`
			} `json:"account"`
		} `json:"user"`
		Modal struct {
			Active interface{} `json:"active"`
			State  struct {
			} `json:"state"`
		} `json:"modal"`
		Notifications struct {
			Notifications        []interface{} `json:"notifications"`
			ListNotifications    []interface{} `json:"listNotifications"`
			ShowShoppableWarning bool          `json:"showShoppableWarning"`
		} `json:"notifications"`
		Mpgs struct {
			FormFieldValidity struct {
				CardNumberValidity  string `json:"cardNumberValidity"`
				ExpiryYearValidity  string `json:"expiryYearValidity"`
				ExpiryMonthValidity string `json:"expiryMonthValidity"`
				CvvValidity         string `json:"cvvValidity"`
			} `json:"formFieldValidity"`
			InitStatus      string      `json:"initStatus"`
			SubmitStatus    string      `json:"submitStatus"`
			SuccessData     interface{} `json:"successData"`
			UnexpectedError bool        `json:"unexpectedError"`
			SaveToProfile   bool        `json:"saveToProfile"`
		} `json:"mpgs"`
		Trolley struct {
			Error                                 interface{}   `json:"error"`
			ItemsBeingUpdated                     []interface{} `json:"itemsBeingUpdated"`
			FailedItemGroups                      []interface{} `json:"failedItemGroups"`
			ResolvedProductIdsFromFailedItemGroup []interface{} `json:"resolvedProductIdsFromFailedItemGroup"`
			StoreID                               string        `json:"storeId"`
			Validation                            struct {
				IsValidating     bool        `json:"isValidating"`
				IsValid          bool        `json:"isValid"`
				ValidationErrors interface{} `json:"validationErrors"`
				Error            interface{} `json:"error"`
				RestrictedItems  interface{} `json:"restrictedItems"`
			} `json:"validation"`
			IsSwappingItems                            bool          `json:"isSwappingItems"`
			UpdateQueue                                []interface{} `json:"updateQueue"`
			UpdateQueueCallbacks                       []interface{} `json:"updateQueueCallbacks"`
			ProcessUpdateQueueImmediately              bool          `json:"processUpdateQueueImmediately"`
			IsProcessUpdateQueueErrorNotificationMuted bool          `json:"isProcessUpdateQueueErrorNotificationMuted"`
			FetchContext                               struct {
			} `json:"fetchContext"`
		} `json:"trolley"`
		Drawer struct {
			Active []interface{} `json:"active"`
			State  struct {
			} `json:"state"`
		} `json:"drawer"`
		ShoppingMethod struct {
			IsEditing        bool `json:"isEditing"`
			DidStoreIDChange bool `json:"didStoreIdChange"`
			State            struct {
			} `json:"state"`
		} `json:"shoppingMethod"`
		EnquiryForms struct {
			Ids      []interface{} `json:"ids"`
			Entities struct {
			} `json:"entities"`
		} `json:"enquiryForms"`
		List struct {
			Error               interface{}   `json:"error"`
			PatchListItemsQueue []interface{} `json:"patchListItemsQueue"`
		} `json:"list"`
		Content struct {
			PageCategoryL1           string        `json:"pageCategoryL1"`
			PageCategoryL2           string        `json:"pageCategoryL2"`
			DisplayFilter            bool          `json:"displayFilter"`
			ExpandFilter             []interface{} `json:"expandFilter"`
			NextLevel                bool          `json:"nextLevel"`
			PageTitle                string        `json:"pageTitle"`
			PageType                 string        `json:"pageType"`
			Breadcrumbs              []interface{} `json:"breadcrumbs"`
			RecipeID                 string        `json:"recipeId"`
			IsDisplayShopIngredients bool          `json:"isDisplayShopIngredients"`
			GlobalUrgencyStrip       struct {
			} `json:"globalUrgencyStrip"`
			RecipeServingSize int  `json:"recipeServingSize"`
			DeliveryMethod    bool `json:"deliveryMethod"`
		} `json:"content"`
		SeoJSONLd struct {
			ComponentJSONLd struct {
			} `json:"componentJsonLd"`
			ShowAsJSONLd bool `json:"showAsJsonLd"`
		} `json:"seoJsonLd"`
		BffAPI struct {
			Queries struct {
			} `json:"queries"`
			Mutations struct {
			} `json:"mutations"`
			Provided struct {
			} `json:"provided"`
			Subscriptions struct {
			} `json:"subscriptions"`
			Config struct {
				Online                    bool   `json:"online"`
				Focused                   bool   `json:"focused"`
				MiddlewareRegistered      bool   `json:"middlewareRegistered"`
				RefetchOnFocus            bool   `json:"refetchOnFocus"`
				RefetchOnReconnect        bool   `json:"refetchOnReconnect"`
				RefetchOnMountOrArgChange bool   `json:"refetchOnMountOrArgChange"`
				KeepUnusedDataFor         int    `json:"keepUnusedDataFor"`
				ReducerPath               string `json:"reducerPath"`
			} `json:"config"`
		} `json:"bffApi"`
		AemAPI struct {
			Queries struct {
			} `json:"queries"`
			Mutations struct {
			} `json:"mutations"`
			Provided struct {
			} `json:"provided"`
			Subscriptions struct {
			} `json:"subscriptions"`
			Config struct {
				Online                    bool   `json:"online"`
				Focused                   bool   `json:"focused"`
				MiddlewareRegistered      bool   `json:"middlewareRegistered"`
				RefetchOnFocus            bool   `json:"refetchOnFocus"`
				RefetchOnReconnect        bool   `json:"refetchOnReconnect"`
				RefetchOnMountOrArgChange bool   `json:"refetchOnMountOrArgChange"`
				KeepUnusedDataFor         int    `json:"keepUnusedDataFor"`
				ReducerPath               string `json:"reducerPath"`
			} `json:"config"`
		} `json:"aemApi"`
		EnquiryFormAPI struct {
			Queries struct {
			} `json:"queries"`
			Mutations struct {
			} `json:"mutations"`
			Provided struct {
			} `json:"provided"`
			Subscriptions struct {
			} `json:"subscriptions"`
			Config struct {
				Online                    bool   `json:"online"`
				Focused                   bool   `json:"focused"`
				MiddlewareRegistered      bool   `json:"middlewareRegistered"`
				RefetchOnFocus            bool   `json:"refetchOnFocus"`
				RefetchOnReconnect        bool   `json:"refetchOnReconnect"`
				RefetchOnMountOrArgChange bool   `json:"refetchOnMountOrArgChange"`
				KeepUnusedDataFor         int    `json:"keepUnusedDataFor"`
				ReducerPath               string `json:"reducerPath"`
			} `json:"config"`
		} `json:"enquiryFormApi"`
		B2BFormsAPI struct {
			Queries struct {
			} `json:"queries"`
			Mutations struct {
			} `json:"mutations"`
			Provided struct {
			} `json:"provided"`
			Subscriptions struct {
			} `json:"subscriptions"`
			Config struct {
				Online                    bool   `json:"online"`
				Focused                   bool   `json:"focused"`
				MiddlewareRegistered      bool   `json:"middlewareRegistered"`
				RefetchOnFocus            bool   `json:"refetchOnFocus"`
				RefetchOnReconnect        bool   `json:"refetchOnReconnect"`
				RefetchOnMountOrArgChange bool   `json:"refetchOnMountOrArgChange"`
				KeepUnusedDataFor         int    `json:"keepUnusedDataFor"`
				ReducerPath               string `json:"reducerPath"`
			} `json:"config"`
		} `json:"b2bFormsApi"`
		RadioComplaintsFormAPI struct {
			Queries struct {
			} `json:"queries"`
			Mutations struct {
			} `json:"mutations"`
			Provided struct {
			} `json:"provided"`
			Subscriptions struct {
			} `json:"subscriptions"`
			Config struct {
				Online                    bool   `json:"online"`
				Focused                   bool   `json:"focused"`
				MiddlewareRegistered      bool   `json:"middlewareRegistered"`
				RefetchOnFocus            bool   `json:"refetchOnFocus"`
				RefetchOnReconnect        bool   `json:"refetchOnReconnect"`
				RefetchOnMountOrArgChange bool   `json:"refetchOnMountOrArgChange"`
				KeepUnusedDataFor         int    `json:"keepUnusedDataFor"`
				ReducerPath               string `json:"reducerPath"`
			} `json:"config"`
		} `json:"radioComplaintsFormApi"`
		PsdsFormAPI struct {
			Queries struct {
			} `json:"queries"`
			Mutations struct {
			} `json:"mutations"`
			Provided struct {
			} `json:"provided"`
			Subscriptions struct {
			} `json:"subscriptions"`
			Config struct {
				Online                    bool   `json:"online"`
				Focused                   bool   `json:"focused"`
				MiddlewareRegistered      bool   `json:"middlewareRegistered"`
				RefetchOnFocus            bool   `json:"refetchOnFocus"`
				RefetchOnReconnect        bool   `json:"refetchOnReconnect"`
				RefetchOnMountOrArgChange bool   `json:"refetchOnMountOrArgChange"`
				KeepUnusedDataFor         int    `json:"keepUnusedDataFor"`
				ReducerPath               string `json:"reducerPath"`
			} `json:"config"`
		} `json:"psdsFormApi"`
		AdobeTargetAPI struct {
			Queries struct {
			} `json:"queries"`
			Mutations struct {
			} `json:"mutations"`
			Provided struct {
			} `json:"provided"`
			Subscriptions struct {
			} `json:"subscriptions"`
			Config struct {
				Online                    bool   `json:"online"`
				Focused                   bool   `json:"focused"`
				MiddlewareRegistered      bool   `json:"middlewareRegistered"`
				RefetchOnFocus            bool   `json:"refetchOnFocus"`
				RefetchOnReconnect        bool   `json:"refetchOnReconnect"`
				RefetchOnMountOrArgChange bool   `json:"refetchOnMountOrArgChange"`
				KeepUnusedDataFor         int    `json:"keepUnusedDataFor"`
				ReducerPath               string `json:"reducerPath"`
			} `json:"config"`
		} `json:"adobeTargetApi"`
		AbandonedTrolleyFormAPI struct {
			Queries struct {
			} `json:"queries"`
			Mutations struct {
			} `json:"mutations"`
			Provided struct {
			} `json:"provided"`
			Subscriptions struct {
			} `json:"subscriptions"`
			Config struct {
				Online                    bool   `json:"online"`
				Focused                   bool   `json:"focused"`
				MiddlewareRegistered      bool   `json:"middlewareRegistered"`
				RefetchOnFocus            bool   `json:"refetchOnFocus"`
				RefetchOnReconnect        bool   `json:"refetchOnReconnect"`
				RefetchOnMountOrArgChange bool   `json:"refetchOnMountOrArgChange"`
				KeepUnusedDataFor         int    `json:"keepUnusedDataFor"`
				ReducerPath               string `json:"reducerPath"`
			} `json:"config"`
		} `json:"abandonedTrolleyFormApi"`
		DigitalGraphQLAPI struct {
			Queries struct {
			} `json:"queries"`
			Mutations struct {
			} `json:"mutations"`
			Provided struct {
			} `json:"provided"`
			Subscriptions struct {
			} `json:"subscriptions"`
			Config struct {
				Online                    bool   `json:"online"`
				Focused                   bool   `json:"focused"`
				MiddlewareRegistered      bool   `json:"middlewareRegistered"`
				RefetchOnFocus            bool   `json:"refetchOnFocus"`
				RefetchOnReconnect        bool   `json:"refetchOnReconnect"`
				RefetchOnMountOrArgChange bool   `json:"refetchOnMountOrArgChange"`
				KeepUnusedDataFor         int    `json:"keepUnusedDataFor"`
				ReducerPath               string `json:"reducerPath"`
			} `json:"config"`
		} `json:"digitalGraphQLApi"`
		NextAPI struct {
			Queries struct {
			} `json:"queries"`
			Mutations struct {
			} `json:"mutations"`
			Provided struct {
			} `json:"provided"`
			Subscriptions struct {
			} `json:"subscriptions"`
			Config struct {
				Online                    bool   `json:"online"`
				Focused                   bool   `json:"focused"`
				MiddlewareRegistered      bool   `json:"middlewareRegistered"`
				RefetchOnFocus            bool   `json:"refetchOnFocus"`
				RefetchOnReconnect        bool   `json:"refetchOnReconnect"`
				RefetchOnMountOrArgChange bool   `json:"refetchOnMountOrArgChange"`
				KeepUnusedDataFor         int    `json:"keepUnusedDataFor"`
				ReducerPath               string `json:"reducerPath"`
			} `json:"config"`
		} `json:"nextApi"`
	} `json:"initialState"`
	NSSP bool `json:"__N_SSP"`
}

type departmentPage struct {
	ID   string
	page int
}

type departmentInfo struct {
	ID               string           `json:"id"`
	Level            int              `json:"level"`
	Name             string           `json:"name"`
	OriginalName     string           `json:"originalName"`
	ProductCount     int              `json:"productCount"`
	SeoToken         string           `json:"seoToken"`
	CatalogGroupView []departmentInfo `json:"catalogGroupView"`
	Image            string           `json:"image"`
	Updated          time.Time
}

type browsePage struct {
	PageProps struct {
		AssetsURL            string `json:"assetsUrl"`
		SentryTraceData      string `json:"_sentryTraceData"`
		SentryBaggage        string `json:"_sentryBaggage"`
		AllProductCategories struct {
			CatalogGroupView    []departmentInfo `json:"catalogGroupView"`
			ExcludedCategoryIds []string         `json:"excludedCategoryIds"`
		} `json:"allProductCategories"`
		InitialState struct {
			User struct {
				Error interface{} `json:"error"`
				Auth  struct {
					Authenticated bool `json:"authenticated"`
				} `json:"auth"`
				Account struct {
					Notifications []interface{} `json:"notifications"`
				} `json:"account"`
			} `json:"user"`
			Modal struct {
				Active interface{} `json:"active"`
				State  struct {
				} `json:"state"`
			} `json:"modal"`
			Notifications struct {
				Notifications        []interface{} `json:"notifications"`
				ListNotifications    []interface{} `json:"listNotifications"`
				ShowShoppableWarning bool          `json:"showShoppableWarning"`
			} `json:"notifications"`
			Mpgs struct {
				FormFieldValidity struct {
					CardNumberValidity  string `json:"cardNumberValidity"`
					ExpiryYearValidity  string `json:"expiryYearValidity"`
					ExpiryMonthValidity string `json:"expiryMonthValidity"`
					CvvValidity         string `json:"cvvValidity"`
				} `json:"formFieldValidity"`
				InitStatus      string      `json:"initStatus"`
				SubmitStatus    string      `json:"submitStatus"`
				SuccessData     interface{} `json:"successData"`
				UnexpectedError bool        `json:"unexpectedError"`
				SaveToProfile   bool        `json:"saveToProfile"`
			} `json:"mpgs"`
			Trolley struct {
				Error                                 interface{}   `json:"error"`
				ItemsBeingUpdated                     []interface{} `json:"itemsBeingUpdated"`
				FailedItemGroups                      []interface{} `json:"failedItemGroups"`
				ResolvedProductIdsFromFailedItemGroup []interface{} `json:"resolvedProductIdsFromFailedItemGroup"`
				StoreID                               string        `json:"storeId"`
				Validation                            struct {
					IsValidating     bool        `json:"isValidating"`
					IsValid          bool        `json:"isValid"`
					ValidationErrors interface{} `json:"validationErrors"`
					Error            interface{} `json:"error"`
					RestrictedItems  interface{} `json:"restrictedItems"`
				} `json:"validation"`
				IsSwappingItems                            bool          `json:"isSwappingItems"`
				UpdateQueue                                []interface{} `json:"updateQueue"`
				UpdateQueueCallbacks                       []interface{} `json:"updateQueueCallbacks"`
				ProcessUpdateQueueImmediately              bool          `json:"processUpdateQueueImmediately"`
				IsProcessUpdateQueueErrorNotificationMuted bool          `json:"isProcessUpdateQueueErrorNotificationMuted"`
				FetchContext                               struct {
				} `json:"fetchContext"`
			} `json:"trolley"`
			Drawer struct {
				Active []interface{} `json:"active"`
				State  struct {
				} `json:"state"`
			} `json:"drawer"`
			ShoppingMethod struct {
				IsEditing        bool `json:"isEditing"`
				DidStoreIDChange bool `json:"didStoreIdChange"`
				State            struct {
				} `json:"state"`
			} `json:"shoppingMethod"`
			EnquiryForms struct {
				Ids      []interface{} `json:"ids"`
				Entities struct {
				} `json:"entities"`
			} `json:"enquiryForms"`
			List struct {
				Error               interface{}   `json:"error"`
				PatchListItemsQueue []interface{} `json:"patchListItemsQueue"`
			} `json:"list"`
			Content struct {
				PageCategoryL1           string        `json:"pageCategoryL1"`
				PageCategoryL2           string        `json:"pageCategoryL2"`
				DisplayFilter            bool          `json:"displayFilter"`
				ExpandFilter             []interface{} `json:"expandFilter"`
				NextLevel                bool          `json:"nextLevel"`
				PageTitle                string        `json:"pageTitle"`
				PageType                 string        `json:"pageType"`
				Breadcrumbs              []interface{} `json:"breadcrumbs"`
				RecipeID                 string        `json:"recipeId"`
				IsDisplayShopIngredients bool          `json:"isDisplayShopIngredients"`
				GlobalUrgencyStrip       struct {
				} `json:"globalUrgencyStrip"`
				RecipeServingSize int  `json:"recipeServingSize"`
				DeliveryMethod    bool `json:"deliveryMethod"`
			} `json:"content"`
			SeoJSONLd struct {
				ComponentJSONLd struct {
				} `json:"componentJsonLd"`
				ShowAsJSONLd bool `json:"showAsJsonLd"`
			} `json:"seoJsonLd"`
			BffAPI struct {
				Queries struct {
				} `json:"queries"`
				Mutations struct {
				} `json:"mutations"`
				Provided struct {
				} `json:"provided"`
				Subscriptions struct {
				} `json:"subscriptions"`
				Config struct {
					Online                    bool   `json:"online"`
					Focused                   bool   `json:"focused"`
					MiddlewareRegistered      bool   `json:"middlewareRegistered"`
					RefetchOnFocus            bool   `json:"refetchOnFocus"`
					RefetchOnReconnect        bool   `json:"refetchOnReconnect"`
					RefetchOnMountOrArgChange bool   `json:"refetchOnMountOrArgChange"`
					KeepUnusedDataFor         int    `json:"keepUnusedDataFor"`
					ReducerPath               string `json:"reducerPath"`
				} `json:"config"`
			} `json:"bffApi"`
			AemAPI struct {
				Queries struct {
				} `json:"queries"`
				Mutations struct {
				} `json:"mutations"`
				Provided struct {
				} `json:"provided"`
				Subscriptions struct {
				} `json:"subscriptions"`
				Config struct {
					Online                    bool   `json:"online"`
					Focused                   bool   `json:"focused"`
					MiddlewareRegistered      bool   `json:"middlewareRegistered"`
					RefetchOnFocus            bool   `json:"refetchOnFocus"`
					RefetchOnReconnect        bool   `json:"refetchOnReconnect"`
					RefetchOnMountOrArgChange bool   `json:"refetchOnMountOrArgChange"`
					KeepUnusedDataFor         int    `json:"keepUnusedDataFor"`
					ReducerPath               string `json:"reducerPath"`
				} `json:"config"`
			} `json:"aemApi"`
			EnquiryFormAPI struct {
				Queries struct {
				} `json:"queries"`
				Mutations struct {
				} `json:"mutations"`
				Provided struct {
				} `json:"provided"`
				Subscriptions struct {
				} `json:"subscriptions"`
				Config struct {
					Online                    bool   `json:"online"`
					Focused                   bool   `json:"focused"`
					MiddlewareRegistered      bool   `json:"middlewareRegistered"`
					RefetchOnFocus            bool   `json:"refetchOnFocus"`
					RefetchOnReconnect        bool   `json:"refetchOnReconnect"`
					RefetchOnMountOrArgChange bool   `json:"refetchOnMountOrArgChange"`
					KeepUnusedDataFor         int    `json:"keepUnusedDataFor"`
					ReducerPath               string `json:"reducerPath"`
				} `json:"config"`
			} `json:"enquiryFormApi"`
			B2BFormsAPI struct {
				Queries struct {
				} `json:"queries"`
				Mutations struct {
				} `json:"mutations"`
				Provided struct {
				} `json:"provided"`
				Subscriptions struct {
				} `json:"subscriptions"`
				Config struct {
					Online                    bool   `json:"online"`
					Focused                   bool   `json:"focused"`
					MiddlewareRegistered      bool   `json:"middlewareRegistered"`
					RefetchOnFocus            bool   `json:"refetchOnFocus"`
					RefetchOnReconnect        bool   `json:"refetchOnReconnect"`
					RefetchOnMountOrArgChange bool   `json:"refetchOnMountOrArgChange"`
					KeepUnusedDataFor         int    `json:"keepUnusedDataFor"`
					ReducerPath               string `json:"reducerPath"`
				} `json:"config"`
			} `json:"b2bFormsApi"`
			RadioComplaintsFormAPI struct {
				Queries struct {
				} `json:"queries"`
				Mutations struct {
				} `json:"mutations"`
				Provided struct {
				} `json:"provided"`
				Subscriptions struct {
				} `json:"subscriptions"`
				Config struct {
					Online                    bool   `json:"online"`
					Focused                   bool   `json:"focused"`
					MiddlewareRegistered      bool   `json:"middlewareRegistered"`
					RefetchOnFocus            bool   `json:"refetchOnFocus"`
					RefetchOnReconnect        bool   `json:"refetchOnReconnect"`
					RefetchOnMountOrArgChange bool   `json:"refetchOnMountOrArgChange"`
					KeepUnusedDataFor         int    `json:"keepUnusedDataFor"`
					ReducerPath               string `json:"reducerPath"`
				} `json:"config"`
			} `json:"radioComplaintsFormApi"`
			PsdsFormAPI struct {
				Queries struct {
				} `json:"queries"`
				Mutations struct {
				} `json:"mutations"`
				Provided struct {
				} `json:"provided"`
				Subscriptions struct {
				} `json:"subscriptions"`
				Config struct {
					Online                    bool   `json:"online"`
					Focused                   bool   `json:"focused"`
					MiddlewareRegistered      bool   `json:"middlewareRegistered"`
					RefetchOnFocus            bool   `json:"refetchOnFocus"`
					RefetchOnReconnect        bool   `json:"refetchOnReconnect"`
					RefetchOnMountOrArgChange bool   `json:"refetchOnMountOrArgChange"`
					KeepUnusedDataFor         int    `json:"keepUnusedDataFor"`
					ReducerPath               string `json:"reducerPath"`
				} `json:"config"`
			} `json:"psdsFormApi"`
			AdobeTargetAPI struct {
				Queries struct {
				} `json:"queries"`
				Mutations struct {
				} `json:"mutations"`
				Provided struct {
				} `json:"provided"`
				Subscriptions struct {
				} `json:"subscriptions"`
				Config struct {
					Online                    bool   `json:"online"`
					Focused                   bool   `json:"focused"`
					MiddlewareRegistered      bool   `json:"middlewareRegistered"`
					RefetchOnFocus            bool   `json:"refetchOnFocus"`
					RefetchOnReconnect        bool   `json:"refetchOnReconnect"`
					RefetchOnMountOrArgChange bool   `json:"refetchOnMountOrArgChange"`
					KeepUnusedDataFor         int    `json:"keepUnusedDataFor"`
					ReducerPath               string `json:"reducerPath"`
				} `json:"config"`
			} `json:"adobeTargetApi"`
			AbandonedTrolleyFormAPI struct {
				Queries struct {
				} `json:"queries"`
				Mutations struct {
				} `json:"mutations"`
				Provided struct {
				} `json:"provided"`
				Subscriptions struct {
				} `json:"subscriptions"`
				Config struct {
					Online                    bool   `json:"online"`
					Focused                   bool   `json:"focused"`
					MiddlewareRegistered      bool   `json:"middlewareRegistered"`
					RefetchOnFocus            bool   `json:"refetchOnFocus"`
					RefetchOnReconnect        bool   `json:"refetchOnReconnect"`
					RefetchOnMountOrArgChange bool   `json:"refetchOnMountOrArgChange"`
					KeepUnusedDataFor         int    `json:"keepUnusedDataFor"`
					ReducerPath               string `json:"reducerPath"`
				} `json:"config"`
			} `json:"abandonedTrolleyFormApi"`
			DigitalGraphQLAPI struct {
				Queries struct {
					GetProductCategoriesStoreIDCOL584 struct {
						Status       string `json:"status"`
						EndpointName string `json:"endpointName"`
						RequestID    string `json:"requestId"`
						OriginalArgs struct {
							StoreID string `json:"storeId"`
						} `json:"originalArgs"`
						StartedTimeStamp int64 `json:"startedTimeStamp"`
						Data             struct {
							ProductCategories struct {
								ExcludedCategoryIds []string `json:"excludedCategoryIds"`
								CatalogGroupView    []struct {
									ID               string `json:"id"`
									Level            int    `json:"level"`
									Name             string `json:"name"`
									OriginalName     string `json:"originalName"`
									ProductCount     int    `json:"productCount"`
									SeoToken         string `json:"seoToken"`
									CatalogGroupView []struct {
										ID               string `json:"id"`
										Level            int    `json:"level"`
										Name             string `json:"name"`
										OriginalName     string `json:"originalName"`
										ProductCount     int    `json:"productCount"`
										SeoToken         string `json:"seoToken"`
										CatalogGroupView []struct {
											ID           string `json:"id"`
											Level        int    `json:"level"`
											Name         string `json:"name"`
											OriginalName string `json:"originalName"`
											ProductCount int    `json:"productCount"`
											SeoToken     string `json:"seoToken"`
										} `json:"catalogGroupView"`
									} `json:"catalogGroupView"`
								} `json:"catalogGroupView"`
							} `json:"productCategories"`
						} `json:"data"`
						FulfilledTimeStamp int64 `json:"fulfilledTimeStamp"`
					} `json:"GetProductCategories({"storeId":"COL:584"})"`
				} `json:"queries"`
				Mutations struct {
				} `json:"mutations"`
				Provided struct {
				} `json:"provided"`
				Subscriptions struct {
					GetProductCategoriesStoreIDCOL584 struct {
						GQcfxHNhsEDvviZ02LN struct {
						} `json:"GQcfxHNhsE-dvviZ02lN_"`
					} `json:"GetProductCategories({"storeId":"COL:584"})"`
				} `json:"subscriptions"`
				Config struct {
					Online                    bool   `json:"online"`
					Focused                   bool   `json:"focused"`
					MiddlewareRegistered      bool   `json:"middlewareRegistered"`
					RefetchOnFocus            bool   `json:"refetchOnFocus"`
					RefetchOnReconnect        bool   `json:"refetchOnReconnect"`
					RefetchOnMountOrArgChange bool   `json:"refetchOnMountOrArgChange"`
					KeepUnusedDataFor         int    `json:"keepUnusedDataFor"`
					ReducerPath               string `json:"reducerPath"`
				} `json:"config"`
			} `json:"digitalGraphQLApi"`
			NextAPI struct {
				Queries struct {
				} `json:"queries"`
				Mutations struct {
				} `json:"mutations"`
				Provided struct {
				} `json:"provided"`
				Subscriptions struct {
				} `json:"subscriptions"`
				Config struct {
					Online                    bool   `json:"online"`
					Focused                   bool   `json:"focused"`
					MiddlewareRegistered      bool   `json:"middlewareRegistered"`
					RefetchOnFocus            bool   `json:"refetchOnFocus"`
					RefetchOnReconnect        bool   `json:"refetchOnReconnect"`
					RefetchOnMountOrArgChange bool   `json:"refetchOnMountOrArgChange"`
					KeepUnusedDataFor         int    `json:"keepUnusedDataFor"`
					ReducerPath               string `json:"reducerPath"`
				} `json:"config"`
			} `json:"nextApi"`
		} `json:"initialState"`
	} `json:"pageProps"`
	InitialState struct {
		User struct {
			Error interface{} `json:"error"`
			Auth  struct {
				Authenticated bool `json:"authenticated"`
			} `json:"auth"`
			Account struct {
				Notifications []interface{} `json:"notifications"`
			} `json:"account"`
		} `json:"user"`
		Modal struct {
			Active interface{} `json:"active"`
			State  struct {
			} `json:"state"`
		} `json:"modal"`
		Notifications struct {
			Notifications        []interface{} `json:"notifications"`
			ListNotifications    []interface{} `json:"listNotifications"`
			ShowShoppableWarning bool          `json:"showShoppableWarning"`
		} `json:"notifications"`
		Mpgs struct {
			FormFieldValidity struct {
				CardNumberValidity  string `json:"cardNumberValidity"`
				ExpiryYearValidity  string `json:"expiryYearValidity"`
				ExpiryMonthValidity string `json:"expiryMonthValidity"`
				CvvValidity         string `json:"cvvValidity"`
			} `json:"formFieldValidity"`
			InitStatus      string      `json:"initStatus"`
			SubmitStatus    string      `json:"submitStatus"`
			SuccessData     interface{} `json:"successData"`
			UnexpectedError bool        `json:"unexpectedError"`
			SaveToProfile   bool        `json:"saveToProfile"`
		} `json:"mpgs"`
		Trolley struct {
			Error                                 interface{}   `json:"error"`
			ItemsBeingUpdated                     []interface{} `json:"itemsBeingUpdated"`
			FailedItemGroups                      []interface{} `json:"failedItemGroups"`
			ResolvedProductIdsFromFailedItemGroup []interface{} `json:"resolvedProductIdsFromFailedItemGroup"`
			StoreID                               string        `json:"storeId"`
			Validation                            struct {
				IsValidating     bool        `json:"isValidating"`
				IsValid          bool        `json:"isValid"`
				ValidationErrors interface{} `json:"validationErrors"`
				Error            interface{} `json:"error"`
				RestrictedItems  interface{} `json:"restrictedItems"`
			} `json:"validation"`
			IsSwappingItems                            bool          `json:"isSwappingItems"`
			UpdateQueue                                []interface{} `json:"updateQueue"`
			UpdateQueueCallbacks                       []interface{} `json:"updateQueueCallbacks"`
			ProcessUpdateQueueImmediately              bool          `json:"processUpdateQueueImmediately"`
			IsProcessUpdateQueueErrorNotificationMuted bool          `json:"isProcessUpdateQueueErrorNotificationMuted"`
			FetchContext                               struct {
			} `json:"fetchContext"`
		} `json:"trolley"`
		Drawer struct {
			Active []interface{} `json:"active"`
			State  struct {
			} `json:"state"`
		} `json:"drawer"`
		ShoppingMethod struct {
			IsEditing        bool `json:"isEditing"`
			DidStoreIDChange bool `json:"didStoreIdChange"`
			State            struct {
			} `json:"state"`
		} `json:"shoppingMethod"`
		EnquiryForms struct {
			Ids      []interface{} `json:"ids"`
			Entities struct {
			} `json:"entities"`
		} `json:"enquiryForms"`
		List struct {
			Error               interface{}   `json:"error"`
			PatchListItemsQueue []interface{} `json:"patchListItemsQueue"`
		} `json:"list"`
		Content struct {
			PageCategoryL1           string        `json:"pageCategoryL1"`
			PageCategoryL2           string        `json:"pageCategoryL2"`
			DisplayFilter            bool          `json:"displayFilter"`
			ExpandFilter             []interface{} `json:"expandFilter"`
			NextLevel                bool          `json:"nextLevel"`
			PageTitle                string        `json:"pageTitle"`
			PageType                 string        `json:"pageType"`
			Breadcrumbs              []interface{} `json:"breadcrumbs"`
			RecipeID                 string        `json:"recipeId"`
			IsDisplayShopIngredients bool          `json:"isDisplayShopIngredients"`
			GlobalUrgencyStrip       struct {
			} `json:"globalUrgencyStrip"`
			RecipeServingSize int  `json:"recipeServingSize"`
			DeliveryMethod    bool `json:"deliveryMethod"`
		} `json:"content"`
		SeoJSONLd struct {
			ComponentJSONLd struct {
			} `json:"componentJsonLd"`
			ShowAsJSONLd bool `json:"showAsJsonLd"`
		} `json:"seoJsonLd"`
		BffAPI struct {
			Queries struct {
			} `json:"queries"`
			Mutations struct {
			} `json:"mutations"`
			Provided struct {
			} `json:"provided"`
			Subscriptions struct {
			} `json:"subscriptions"`
			Config struct {
				Online                    bool   `json:"online"`
				Focused                   bool   `json:"focused"`
				MiddlewareRegistered      bool   `json:"middlewareRegistered"`
				RefetchOnFocus            bool   `json:"refetchOnFocus"`
				RefetchOnReconnect        bool   `json:"refetchOnReconnect"`
				RefetchOnMountOrArgChange bool   `json:"refetchOnMountOrArgChange"`
				KeepUnusedDataFor         int    `json:"keepUnusedDataFor"`
				ReducerPath               string `json:"reducerPath"`
			} `json:"config"`
		} `json:"bffApi"`
		AemAPI struct {
			Queries struct {
			} `json:"queries"`
			Mutations struct {
			} `json:"mutations"`
			Provided struct {
			} `json:"provided"`
			Subscriptions struct {
			} `json:"subscriptions"`
			Config struct {
				Online                    bool   `json:"online"`
				Focused                   bool   `json:"focused"`
				MiddlewareRegistered      bool   `json:"middlewareRegistered"`
				RefetchOnFocus            bool   `json:"refetchOnFocus"`
				RefetchOnReconnect        bool   `json:"refetchOnReconnect"`
				RefetchOnMountOrArgChange bool   `json:"refetchOnMountOrArgChange"`
				KeepUnusedDataFor         int    `json:"keepUnusedDataFor"`
				ReducerPath               string `json:"reducerPath"`
			} `json:"config"`
		} `json:"aemApi"`
		EnquiryFormAPI struct {
			Queries struct {
			} `json:"queries"`
			Mutations struct {
			} `json:"mutations"`
			Provided struct {
			} `json:"provided"`
			Subscriptions struct {
			} `json:"subscriptions"`
			Config struct {
				Online                    bool   `json:"online"`
				Focused                   bool   `json:"focused"`
				MiddlewareRegistered      bool   `json:"middlewareRegistered"`
				RefetchOnFocus            bool   `json:"refetchOnFocus"`
				RefetchOnReconnect        bool   `json:"refetchOnReconnect"`
				RefetchOnMountOrArgChange bool   `json:"refetchOnMountOrArgChange"`
				KeepUnusedDataFor         int    `json:"keepUnusedDataFor"`
				ReducerPath               string `json:"reducerPath"`
			} `json:"config"`
		} `json:"enquiryFormApi"`
		B2BFormsAPI struct {
			Queries struct {
			} `json:"queries"`
			Mutations struct {
			} `json:"mutations"`
			Provided struct {
			} `json:"provided"`
			Subscriptions struct {
			} `json:"subscriptions"`
			Config struct {
				Online                    bool   `json:"online"`
				Focused                   bool   `json:"focused"`
				MiddlewareRegistered      bool   `json:"middlewareRegistered"`
				RefetchOnFocus            bool   `json:"refetchOnFocus"`
				RefetchOnReconnect        bool   `json:"refetchOnReconnect"`
				RefetchOnMountOrArgChange bool   `json:"refetchOnMountOrArgChange"`
				KeepUnusedDataFor         int    `json:"keepUnusedDataFor"`
				ReducerPath               string `json:"reducerPath"`
			} `json:"config"`
		} `json:"b2bFormsApi"`
		RadioComplaintsFormAPI struct {
			Queries struct {
			} `json:"queries"`
			Mutations struct {
			} `json:"mutations"`
			Provided struct {
			} `json:"provided"`
			Subscriptions struct {
			} `json:"subscriptions"`
			Config struct {
				Online                    bool   `json:"online"`
				Focused                   bool   `json:"focused"`
				MiddlewareRegistered      bool   `json:"middlewareRegistered"`
				RefetchOnFocus            bool   `json:"refetchOnFocus"`
				RefetchOnReconnect        bool   `json:"refetchOnReconnect"`
				RefetchOnMountOrArgChange bool   `json:"refetchOnMountOrArgChange"`
				KeepUnusedDataFor         int    `json:"keepUnusedDataFor"`
				ReducerPath               string `json:"reducerPath"`
			} `json:"config"`
		} `json:"radioComplaintsFormApi"`
		PsdsFormAPI struct {
			Queries struct {
			} `json:"queries"`
			Mutations struct {
			} `json:"mutations"`
			Provided struct {
			} `json:"provided"`
			Subscriptions struct {
			} `json:"subscriptions"`
			Config struct {
				Online                    bool   `json:"online"`
				Focused                   bool   `json:"focused"`
				MiddlewareRegistered      bool   `json:"middlewareRegistered"`
				RefetchOnFocus            bool   `json:"refetchOnFocus"`
				RefetchOnReconnect        bool   `json:"refetchOnReconnect"`
				RefetchOnMountOrArgChange bool   `json:"refetchOnMountOrArgChange"`
				KeepUnusedDataFor         int    `json:"keepUnusedDataFor"`
				ReducerPath               string `json:"reducerPath"`
			} `json:"config"`
		} `json:"psdsFormApi"`
		AdobeTargetAPI struct {
			Queries struct {
			} `json:"queries"`
			Mutations struct {
			} `json:"mutations"`
			Provided struct {
			} `json:"provided"`
			Subscriptions struct {
			} `json:"subscriptions"`
			Config struct {
				Online                    bool   `json:"online"`
				Focused                   bool   `json:"focused"`
				MiddlewareRegistered      bool   `json:"middlewareRegistered"`
				RefetchOnFocus            bool   `json:"refetchOnFocus"`
				RefetchOnReconnect        bool   `json:"refetchOnReconnect"`
				RefetchOnMountOrArgChange bool   `json:"refetchOnMountOrArgChange"`
				KeepUnusedDataFor         int    `json:"keepUnusedDataFor"`
				ReducerPath               string `json:"reducerPath"`
			} `json:"config"`
		} `json:"adobeTargetApi"`
		AbandonedTrolleyFormAPI struct {
			Queries struct {
			} `json:"queries"`
			Mutations struct {
			} `json:"mutations"`
			Provided struct {
			} `json:"provided"`
			Subscriptions struct {
			} `json:"subscriptions"`
			Config struct {
				Online                    bool   `json:"online"`
				Focused                   bool   `json:"focused"`
				MiddlewareRegistered      bool   `json:"middlewareRegistered"`
				RefetchOnFocus            bool   `json:"refetchOnFocus"`
				RefetchOnReconnect        bool   `json:"refetchOnReconnect"`
				RefetchOnMountOrArgChange bool   `json:"refetchOnMountOrArgChange"`
				KeepUnusedDataFor         int    `json:"keepUnusedDataFor"`
				ReducerPath               string `json:"reducerPath"`
			} `json:"config"`
		} `json:"abandonedTrolleyFormApi"`
		DigitalGraphQLAPI struct {
			Queries struct {
			} `json:"queries"`
			Mutations struct {
			} `json:"mutations"`
			Provided struct {
			} `json:"provided"`
			Subscriptions struct {
			} `json:"subscriptions"`
			Config struct {
				Online                    bool   `json:"online"`
				Focused                   bool   `json:"focused"`
				MiddlewareRegistered      bool   `json:"middlewareRegistered"`
				RefetchOnFocus            bool   `json:"refetchOnFocus"`
				RefetchOnReconnect        bool   `json:"refetchOnReconnect"`
				RefetchOnMountOrArgChange bool   `json:"refetchOnMountOrArgChange"`
				KeepUnusedDataFor         int    `json:"keepUnusedDataFor"`
				ReducerPath               string `json:"reducerPath"`
			} `json:"config"`
		} `json:"digitalGraphQLApi"`
		NextAPI struct {
			Queries struct {
			} `json:"queries"`
			Mutations struct {
			} `json:"mutations"`
			Provided struct {
			} `json:"provided"`
			Subscriptions struct {
			} `json:"subscriptions"`
			Config struct {
				Online                    bool   `json:"online"`
				Focused                   bool   `json:"focused"`
				MiddlewareRegistered      bool   `json:"middlewareRegistered"`
				RefetchOnFocus            bool   `json:"refetchOnFocus"`
				RefetchOnReconnect        bool   `json:"refetchOnReconnect"`
				RefetchOnMountOrArgChange bool   `json:"refetchOnMountOrArgChange"`
				KeepUnusedDataFor         int    `json:"keepUnusedDataFor"`
				ReducerPath               string `json:"reducerPath"`
			} `json:"config"`
		} `json:"nextApi"`
	} `json:"initialState"`
	NSSP bool `json:"__N_SSP"`
}
