package rtm // import "suy.io/bots/slack/api/rtm"

// ffjson: noencoder
type Message struct {
	Channel    string `json:"channel" url:"channel"`
	SourceTeam string `json:"source_team" url:"source_team"`
	Team       string `json:"team" url:"team"`
	Text       string `json:"text" url:"text"`
	ThreadTs   string `json:"thread_ts" url:"thread_ts"`
	Ts         string `json:"ts" url:"ts"`
	User       string `json:"user" url:"user"`
}

// ffjson: noencoder
type ChannelJoinMessage struct {
	User        string `json:"user" url:"user"`
	Inviter     string `json:"inviter" url:"inviter"`
	UserProfile struct {
		AvatarHash        string      `json:"avatar_hash" url:"avatar_hash"`
		Image72           string      `json:"image_72" url:"image_72"`
		FirstName         interface{} `json:"first_name" url:"first_name"`
		RealName          string      `json:"real_name" url:"real_name"`
		DisplayName       string      `json:"display_name" url:"display_name"`
		Team              string      `json:"team" url:"team"`
		Name              string      `json:"name" url:"name"`
		IsRestricted      bool        `json:"is_restricted" url:"is_restricted"`
		IsUltraRestricted bool        `json:"is_ultra_restricted" url:"is_ultra_restricted"`
	} `json:"user_profile" url:"user_profile"`
	Type    string `json:"type" url:"type"`
	Subtype string `json:"subtype" url:"subtype"`
	Team    string `json:"team" url:"team"`
	Text    string `json:"text" url:"text"`
	Channel string `json:"channel" url:"channel"`
	EventTs string `json:"event_ts" url:"event_ts"`
	Ts      string `json:"ts" url:"ts"`
}

// ffjson: noencoder
type UserChannelJoinMessage struct {
	Type        string `json:"type" url:"type"`
	User        string `json:"user" url:"user"`
	Channel     string `json:"channel" url:"channel"`
	ChannelType string `json:"channel_type" url:"channel_type"`
	Team        string `json:"team" url:"team"`
	Inviter     string `json:"inviter" url:"inviter"`
	EventTs     string `json:"event_ts" url:"event_ts"`
	Ts          string `json:"ts" url:"ts"`
}

// ffjson: noencoder
type GroupJoinMessage struct {
	Channel struct {
		ID                 string        `json:"id" url:"id"`
		Name               string        `json:"name" url:"name"`
		IsChannel          bool          `json:"is_channel" url:"is_channel"`
		IsGroup            bool          `json:"is_group" url:"is_group"`
		IsIm               bool          `json:"is_im" url:"is_im"`
		Created            int           `json:"created" url:"created"`
		IsArchived         bool          `json:"is_archived" url:"is_archived"`
		IsGeneral          bool          `json:"is_general" url:"is_general"`
		Unlinked           int           `json:"unlinked" url:"unlinked"`
		NameNormalized     string        `json:"name_normalized" url:"name_normalized"`
		IsShared           bool          `json:"is_shared" url:"is_shared"`
		Creator            string        `json:"creator" url:"creator"`
		IsExtShared        bool          `json:"is_ext_shared" url:"is_ext_shared"`
		IsOrgShared        bool          `json:"is_org_shared" url:"is_org_shared"`
		PendingShared      []interface{} `json:"pending_shared" url:"pending_shared"`
		IsPendingExtShared bool          `json:"is_pending_ext_shared" url:"is_pending_ext_shared"`
		IsMember           bool          `json:"is_member" url:"is_member"`
		IsPrivate          bool          `json:"is_private" url:"is_private"`
		IsMpim             bool          `json:"is_mpim" url:"is_mpim"`
		LastRead           string        `json:"last_read" url:"last_read"`
		Latest             interface{}   `json:"latest" url:"latest"`
		UnreadCount        int           `json:"unread_count" url:"unread_count"`
		UnreadCountDisplay int           `json:"unread_count_display" url:"unread_count_display"`
		IsOpen             bool          `json:"is_open" url:"is_open"`
		Members            []string      `json:"members" url:"members"`
		Topic              struct {
			Value   string `json:"value" url:"value"`
			Creator string `json:"creator" url:"creator"`
			LastSet int    `json:"last_set" url:"last_set"`
		} `json:"topic" url:"topic"`
		Purpose struct {
			Value   string `json:"value" url:"value"`
			Creator string `json:"creator" url:"creator"`
			LastSet int    `json:"last_set" url:"last_set"`
		} `json:"purpose" url:"purpose"`
	} `json:"channel" url:"channel"`
}

//go:generate ffjson $GOFILE
