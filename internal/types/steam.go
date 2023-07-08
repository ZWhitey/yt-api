package types

type Inventory struct {
	Assets []struct {
		Appid      int    `json:"appid"`
		Contextid  string `json:"contextid"`
		Assetid    string `json:"assetid"`
		Classid    string `json:"classid"`
		Instanceid string `json:"instanceid"`
		Amount     string `json:"amount"`
	} `json:"assets"`
	Descriptions []struct {
		Appid            int    `json:"appid"`
		Classid          string `json:"classid"`
		Instanceid       string `json:"instanceid"`
		Currency         int    `json:"currency"`
		Background_color string `json:"background_color"`
		Icon_url         string `json:"icon_url"`
		Icon_url_large   string `json:"icon_url_large"`
		Descriptions     []struct {
			Value string `json:"value"`
			Color string `json:"color,omitempty"`
		} `json:"descriptions"`
		Tradable int `json:"tradable"`
		Actions  []struct {
			Link string `json:"link"`
			Name string `json:"name"`
		} `json:"actions"`
		Name                          string `json:"name"`
		Name_color                    string `json:"name_color"`
		Type                          string `json:"type"`
		Market_name                   string `json:"market_name"`
		Market_hash_name              string `json:"market_hash_name"`
		Commodity                     int    `json:"commodity"`
		Market_tradable_restriction   int    `json:"market_tradable_restriction"`
		Market_marketable_restriction int    `json:"market_marketable_restriction"`
		Marketable                    int    `json:"marketable"`
		Tags                          []struct {
			Category                string `json:"category"`
			Internal_name           string `json:"internal_name"`
			Localized_category_name string `json:"localized_category_name"`
			Localized_tag_name      string `json:"localized_tag_name"`
			Color                   string `json:"color"`
		} `json:"tags"`
	} `json:"descriptions"`
	More_items            int    `json:"more_items"`
	Last_assetid          string `json:"last_assetid"`
	Total_inventory_count int    `json:"total_inventory_count"`
	Success               int    `json:"success"`
	Rwgrsn                int    `json:"rwgrsn"`
}

type ProfileResponse struct {
	Response struct {
		Players []struct {
			SteamID             string `json:"steamid"`
			CommunityVisibility int    `json:"communityvisibilitystate"`
			ProfileState        int    `json:"profilestate"`
			PersonaName         string `json:"personaname"`
			ProfileURL          string `json:"profileurl"`
			Avatar              string `json:"avatar"`
			AvatarMedium        string `json:"avatarmedium"`
			AvatarFull          string `json:"avatarfull"`
			AvatarHash          string `json:"avatarhash"`
			PersonaState        int    `json:"personastate"`
			RealName            string `json:"realname"`
			PrimaryClanID       string `json:"primaryclanid"`
			TimeCreated         int    `json:"timecreated"`
			PersonaStateFlags   int    `json:"personastateflags"`
			LoccountryCode      string `json:"loccountrycode"`
			LocStateCode        string `json:"locstatecode"`
			LoccityID           int    `json:"loccityid"`
		} `json:"players"`
	} `json:"response"`
}

type MarketItem struct {
	Success           int    `json:"success"`
	SellOrderTable    string `json:"sell_order_table"`
	SellOrderSummary  string `json:"sell_order_summary"`
	BuyOrderTable     string `json:"buy_order_table"`
	BuyOrderSummary   string `json:"buy_order_summary"`
	HighestBuyOrder   string `json:"highest_buy_order"`
	LowestSellOrder   string `json:"lowest_sell_order"`
	GraphMaxY int `json:"graph_max_y"`
	GraphMinY float64 `json:"graph_min_y"`
	GraphMaxX float64 `json:"graph_max_x"`
	PricePrefix string `json:"price_prefix"`
	PriceSuffix string `json:"price_suffix"`
}
