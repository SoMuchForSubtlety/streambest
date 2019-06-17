streambest is designed to automatically pick tracks from a file with multiple audio and video tracks and then stream those with ffmpeg.  
It will pick the audio track with the language specified in the config and the video track with the highest resolution.

## Usage
Save the `sample-streambest-config.json` as `streambest-config.json` and modify it so it fits your needs, then run streambest like this

```
$ streambest --source {media}
```

You can use any source ffmpeg can normally use.  
You might need to change the ffmpeg command in the config a bit for your usecase (for example to downmix 5.1 audio to stereo).
