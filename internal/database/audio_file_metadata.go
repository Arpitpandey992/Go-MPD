package database

type Tags struct {
	Title       []string            `json:"title"`
	Album       []string            `json:"album"`
	Artist      []string            `json:"artist"`
	AlbumArtist []string            `json:"album_artist"`
	DiscNumber  *int                `json:"disc_number,omitempty"`
	TotalDiscs  *int                `json:"total_discs,omitempty"`
	TrackNumber *int                `json:"track_number,omitempty"`
	TotalTracks *int                `json:"total_tracks,omitempty"`
	Comment     []string            `json:"comment"`
	Date        *string             `json:"date,omitempty"`
	Catalog     []string            `json:"catalog"`
	Barcode     []string            `json:"barcode"`
	DiscName    []string            `json:"disc_name"`
	CustomTags  map[string][]string `json:"custom_tags"`
	Pictures    []string            `json:"pictures"`
	Extension   string              `json:"extension"`

	// Currently not supported by unigen
	Genre          *string `json:"genre,omitempty"`
	Duration       *string `json:"duration,omitempty"`
	Arranger       *string `json:"arranger,omitempty"`
	Author         *string `json:"author,omitempty"`
	Bpm            *string `json:"bpm,omitempty"`
	Composer       *string `json:"composer,omitempty"`
	Conductor      *string `json:"conductor,omitempty"`
	Copyright      *string `json:"copyright,omitempty"`
	EncodedBy      *string `json:"encoded_by,omitempty"`
	Grouping       *string `json:"grouping,omitempty"`
	Isrc           *string `json:"isrc,omitempty"`
	Language       *string `json:"language,omitempty"`
	Lyricist       *string `json:"lyricist,omitempty"`
	Lyrics         *string `json:"lyrics,omitempty"`
	Media          *string `json:"media,omitempty"`
	OriginalAlbum  *string `json:"original_album,omitempty"`
	OriginalArtist *string `json:"original_artist,omitempty"`
	OriginalDate   *string `json:"original_date,omitempty"`
	Part           *string `json:"part,omitempty"`
	Performer      *string `json:"performer,omitempty"`
	Publisher      *string `json:"publisher,omitempty"`
	Remixer        *string `json:"remixer,omitempty"`
	Subtitle       *string `json:"subtitle,omitempty"`
	Website        *string `json:"website,omitempty"`
}

type MediaInfo struct {
	SampleRate    *int    `json:"sample_rate,omitempty"`
	Channels      *int    `json:"channels,omitempty"`
	Bitrate       *int    `json:"bitrate,omitempty"`
	BitsPerSample *int    `json:"bits_per_sample,omitempty"`
	Codec         *string `json:"codec,omitempty"`
}

type AudioFileMetadata struct {
	FileName  string    `json:"file_name"`
	FilePath  string    `json:"file_path"`
	Extension string    `json:"extension"`
	Tags      Tags      `json:"tags"`
	MediaInfo MediaInfo `json:"media_info"`
}
