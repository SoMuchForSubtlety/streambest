package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"

	"golang.org/x/text/language"
)

var media = flag.String("source", "", "the URL to the mediayou want to stream")

type config struct {
	Ingest       string   `json:"ingest"`
	Key          string   `json:"key"`
	PrefLanguage string   `json:"pref_language"`
	Command      []string `json:"command"`
	Language     language.Base
	WantFx       bool
}

type streamInfo struct {
	Streams []struct {
		Index          int    `json:"index"`
		CodecName      string `json:"codec_name"`
		CodecLongName  string `json:"codec_long_name"`
		Profile        string `json:"profile"`
		CodecType      string `json:"codec_type"`
		CodecTimeBase  string `json:"codec_time_base"`
		CodecTagString string `json:"codec_tag_string"`
		CodecTag       string `json:"codec_tag"`
		SampleFmt      string `json:"sample_fmt,omitempty"`
		SampleRate     string `json:"sample_rate,omitempty"`
		Channels       int    `json:"channels,omitempty"`
		ChannelLayout  string `json:"channel_layout,omitempty"`
		BitsPerSample  int    `json:"bits_per_sample,omitempty"`
		RFrameRate     string `json:"r_frame_rate"`
		AvgFrameRate   string `json:"avg_frame_rate"`
		TimeBase       string `json:"time_base"`
		StartPts       int    `json:"start_pts"`
		StartTime      string `json:"start_time"`
		BitRate        string `json:"bit_rate,omitempty"`
		Disposition    struct {
			Default         int `json:"default"`
			Dub             int `json:"dub"`
			Original        int `json:"original"`
			Comment         int `json:"comment"`
			Lyrics          int `json:"lyrics"`
			Karaoke         int `json:"karaoke"`
			Forced          int `json:"forced"`
			HearingImpaired int `json:"hearing_impaired"`
			VisualImpaired  int `json:"visual_impaired"`
			CleanEffects    int `json:"clean_effects"`
			AttachedPic     int `json:"attached_pic"`
			TimedThumbnails int `json:"timed_thumbnails"`
		} `json:"disposition"`
		Tags struct {
			VariantBitrate                                     string `json:"variant_bitrate"`
			ID3V2PrivComAppleStreamingTransportStreamTimestamp string `json:"id3v2_priv.com.apple.streaming.transportStreamTimestamp"`
			Language                                           string `json:"language"`
			Comment                                            string `json:"comment"`
		} `json:"tags"`
		Width              int    `json:"width,omitempty"`
		Height             int    `json:"height,omitempty"`
		CodedWidth         int    `json:"coded_width,omitempty"`
		CodedHeight        int    `json:"coded_height,omitempty"`
		HasBFrames         int    `json:"has_b_frames,omitempty"`
		SampleAspectRatio  string `json:"sample_aspect_ratio,omitempty"`
		DisplayAspectRatio string `json:"display_aspect_ratio,omitempty"`
		PixFmt             string `json:"pix_fmt,omitempty"`
		Level              int    `json:"level,omitempty"`
		ChromaLocation     string `json:"chroma_location,omitempty"`
		Refs               int    `json:"refs,omitempty"`
		IsAvc              string `json:"is_avc,omitempty"`
		NalLengthSize      string `json:"nal_length_size,omitempty"`
		BitsPerRawSample   string `json:"bits_per_raw_sample,omitempty"`
	} `json:"streams"`
	Format struct {
		Filename       string `json:"filename"`
		NbStreams      int    `json:"nb_streams"`
		NbPrograms     int    `json:"nb_programs"`
		FormatName     string `json:"format_name"`
		FormatLongName string `json:"format_long_name"`
		StartTime      string `json:"start_time"`
		Duration       string `json:"duration"`
		Size           string `json:"size"`
		BitRate        string `json:"bit_rate"`
		ProbeScore     int    `json:"probe_score"`
	} `json:"format"`
}

func main() {
	flag.Parse()

	if *media == "" {
		log.Fatal("no media specified")
	}

	var cfg config
	configFile, err := os.Open("config.json")
	defer configFile.Close()
	if err != nil {
		log.Fatal(err.Error())
	}
	jsonParser := json.NewDecoder(configFile)

	cfg.PrefLanguage = "eng"

	err = jsonParser.Decode(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	if cfg.Ingest[len(cfg.Ingest)-1:] != "/" {
		cfg.Ingest += "/"
	}
	if len(cfg.Command) == 0 {
		log.Fatal("no commad provided")
	}

	cfg.Language, err = language.ParseBase(cfg.PrefLanguage)
	if err != nil {
		cfg.WantFx = true
	}

	cmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", *media)
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}

	var info streamInfo
	json.Unmarshal(out, &info)
	if err != nil {
		log.Fatal(err)
	}

	backupLang, _ := language.ParseBase("en")
	var foundPrimaryAudio bool
	audioIndex := 0
	audioBackupIndex := 0

	bestHorizRes := 0
	videoIndex := 0

	for _, stream := range info.Streams {
		if stream.CodecType == "audio" && !foundPrimaryAudio {
			foundTag, err := language.ParseBase(stream.Tags.Language)
			if err != nil && cfg.WantFx {
				foundPrimaryAudio = true
				audioIndex = stream.Index
			} else if foundTag == cfg.Language {
				foundPrimaryAudio = true
				audioIndex = stream.Index
			} else if foundTag == backupLang {
				audioBackupIndex = stream.Index
			}
		} else if stream.CodecType == "video" {
			if bestHorizRes < stream.Width {
				bestHorizRes = stream.Width
				videoIndex = stream.Index
			}
		}
	}

	if !foundPrimaryAudio {
		audioIndex = audioBackupIndex
	}

	for i, arg := range cfg.Command {
		if arg == "$video" {
			cfg.Command[i] = fmt.Sprintf("0:%d", videoIndex)
		} else if arg == "$audio" {
			cfg.Command[i] = fmt.Sprintf("0:%d", audioIndex)
		} else if arg == "$target" {
			cfg.Command[i] = cfg.Ingest + cfg.Key
		} else if arg == "$media" {
			cfg.Command[i] = *media
		}
	}

	fmt.Printf("%v ", cfg.Command)
	fmt.Println()
	cmd = exec.Command(cfg.Command[0], cfg.Command[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	cmd.Start()
}
