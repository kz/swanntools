# swanntools
Capture tools for the Swann DVR4-1200 DVR, inspired by [meatheadmike/swanntools](https://github.com/meatheadmike/swanntools) which was developed for DVR8-2600. Whereas the latter's mobile stream script is compatible with the DVR4-1200, the media script (featuring much higher quality) does not.

This repository is part of a long-term project to build a more secure system piggybacking off the DVR4-1200 with cloud backups and much better web/mobile clients.

## Roadmap

Current task: Conduct research into authentication using Wireshark

- [ ] Create a Python script which can authenticate with the DVR via its media protocol
- [ ] Add streaming of cameras to the script
- [ ] Plan a method to stream the H264 stream to AWS/Azure in order to transcode, store and stream video

## Research
