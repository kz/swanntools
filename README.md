# swanntools
Capture tools for the Swann DVR4-1200 DVR (also known as DVR04B and DM70-D, manufactured by RaySharp), inspired by [meatheadmike/swanntools](https://github.com/meatheadmike/swanntools) which was developed for DVR8-2600. Whereas the latter's mobile stream script is compatible with the DVR4-1200, the media script (featuring much higher quality) does not.

This repository is part of a long-term project to build a more secure system piggybacking off the DVR4-1200 with cloud backups and much better web/mobile clients.

## Structure

The below directory listing explains the structure of this repository.

```
.
├── src                           # Source files
│   ├── auth                      # Miscellaneous code to test the web panel login protocol of the DVR,
│   │   │                         # made redundant as the DVR authenticates camera streaming separately
│   │   ├── authenticate.go       
│   │   │
│   │   └── authenticate_test.go
│   └── client                    # Retrieves and forwards DVR camera streams to the server
│       ├── client.go             # Handles forwarding of streams to server
│       ├── main.go               # Command line point of entry
│       └── stream.go             # Handles connection and receiving streams from the DVR
├── .gitignore
├── LICENSE.md
└── README.md
```

## Known Bugs

- Pressing Ctrl + C does not terminate main.exe

## Roadmap

- [X] Create a Go script which can authenticate with the DVR via its media protocol
- [ ] Add streaming of cameras to the script
    - [X] Receive a continuous stream of a single channel
    - [X] Receive a continuous stream of multiple channels
    - [ ] Implement a TCP proxy
        - [X] Implement a client
        - [ ] Implement a server
- [ ] Plan a method to stream the H264 stream to ~~AWS/~~Azure in order to transcode, store and stream video
    - Plan must involve creation of an interface to allow creation of an AWS script
    - The [Azure/azure-sdk-for-go](https://github.com/Azure/azure-sdk-for-go) can be used to upload files to cold blob storage
- [ ] Create a web client to consume the H264 stream
    - React is the framework of choice (can also easily port to mobile clients using React Native)
    - [nareix/joy4](https://github.com/nareix/joy4) could be used as a streaming server
    - [keroserene/go-webrtc](https://github.com/keroserene/go-webrtc) looks useful as a WebRTC library
- [ ] Create a script to transfer DVR HDD recordings (which are of much higher quality than live streams) to cold storage just in case the extra resolution is required (otherwise can fall back to stored live streams, e.g., if DVR HDD fails)

## Research Journal

The purpose of this section is to detail the network messages sent in Swann's web client.

The following observations can be reproduced using Wireshark captures with the capture filter `host [DVR IP] and port [MEDIA PORT (9000)]` while using the web client (default port 85). All communication with the DVR is made over TCP.

Useful tools:

- A hex editor such as [hexed.it](hexed.it) will be very useful for interpreting the hexdumps below. In the case of hexed.it, remember to substitute the dummy characters (e.g., `X`) for real hex values and when pasting the data, specify that the data should be interpreted as hexadecimal.
- You will also find that an ASCII to Hex converter may be useful. I recommend [asciitohex.com](http://www.asciitohex.com/).
- The command `echo '[INPUT]' | xxd -r -p | nc [DVR IP] [MEDIA PORT]` will be very useful as you can immediately reproduce the below messages by directly replacing `[INPUT]` with the quoted messages.
- If you have VLC installed, `vlc --demux=h264 [FILE]` can play raw camera streams

The below research is in chronological order. Interesting revelations are made throughout.

### Authentication (2017-01-10 to 2017-01-12)

The web client's authentication stages are as follows:

1. Send a message establishing an intent to authenticate
2. Receive a message acknowledging your intent
3. Send a message with authentication data
4. Receive a response with the outcome of the authentication

---

To establish an intent to authenticate, send the following message:

>00000000000000000000010000000a**XX**000000292300000000001c010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000

Notice the dummy characters `XX`. These represent arbitrary hex values (such as '7b') which will be needed in the authentication stage. 

---

You will then receive a response message acknowledging your intent:

>000000010000000a**XX**000000292300000000001c010000000100961200000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000

Notice that you see `XX` again. `XX` will be the same hex values as you provided in the first step (e.g., if you sent your intent message by replacing `XX` with `7b`, the response will have `7b` in place of `XX`).

---

Next, send an authentication message:

>000000000000000000000100000019**YY**0000000000000000000054**UUUUUUUUUU**000000000000000000000000000000000000000000000000000000**PPPPPPPPPPP**000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000

Notice that we have introduced three new dummy characters, `Y`, `U` and `P`:

- `YY` is related to `XX`. Whatever you set `XX` as, `YY` will be equal to `XX` + 1 (using hexadecimal arithmetic). For example, if `XX` was `7b`, `YY` would be `7c`. Likewise, if `XX` was `89`, `YY` would be `8a`.
- `U` and `P` are the hex equivalents for your ASCII username and password respectively. In the above example, the username is five ASCII characters long and the password is six ASCII characters long. For example, if your username is `admin`, your run of `U` values would be `61646d696e`. Likewise, if your password is `passwd`, your run of `P` values would be `706173737764`. If either values is longer than six characters, then simply overwrite the succeeding `00`s with more `UU`s/`PP`s. 

(Side note: we can see from above that our authentication details is sent over a TCP stream. Furthermore, users are constrained to a small character set with a low character limit for their usernames/passwords. This was a major factor of why I decided to start this project.)

---

Finally, the DVR will return one of two responses of length 8:

- If authentication succeeded, it will return `08 00 00 00 02 00 00 00`. 
- If authentication failed, it will return `08 00 00 00 FF FF FF FF`.

---

Observing the above process multiple times, you will find that the value of `XX` (and therefore `YY`) sent by your web client will increase by one or more every time you log in. I can only guess the possible reasons. Perhaps it may be to track sessions? However, it appears that from running the [src/authenticate.go](src/authenticate.go) script for multiple intent values, the **the intent value does not need to increment** and can stay constant.

### DVR Settings (and lack of authentication) (2017-01-13)

Immediately after authentication, the web client sends a message to retrieve the DVR settings (containing MAC address, firmware version, SMTP details (inc. password), admin/user login details):

>00000000000000000000010000000e020000000000000000000014000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000

The DVR responds with roughly twelve kilobytes of information.

After replaying the request without any authentication, I realised that you **do not need to login to get the same response** when you send this packet. I confirmed this by connecting to my media server port from a remote VPS, sending the packet via netcat and seeing the response. I immediately blocked my media port after.

It appears that the only purpose for the authentication protocol is to slow down dumb attackers trying to bruteforce the web client. (Note the use of "slow down", because a maximum of six alphanumeric characters will take no time to bruteforce.) As a result, I am not expecting H264 streaming to require authentication. However, this is good because there will be less code to write, and the purpose of this is to have a Raspberry Pi act as a proxy with proper authentication.

---

### Camera Stream (2017-01-25)

Following on from the last section, I was incorrect in assuming that H264 streaming does not require authentication. That being said, an attacker can use the above method to get the credentials first, defeating the point of the below authentication process.

The following message returns a camera stream:

>00000000000000000000010000000300000000000000000000006800000001000000100000000**N**0000000100000000**UUUUUUUUUU**000000000000000000000000000001000000000000010124000000**PPPPPPPPPPPP**00009cc9c805000000000400010004000000a8c9c80500000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000

Notice the following variables:
- **U** and **P** are as before (username and password)
- **N** is your camera channel. This can be `1`, `2`, `4`, `8`, for channels 1, 2, 3 and 4 respectively

If the authentication fails, the DVR will return `08 00 00 00 04 00 00 00`. 

If the authentication succeeds, the following messages are sent in the following order:
- 8 bytes: `1000000000000000`
- A 1460 byte packet containing the camera stream, with the characters `MDVR96NT` near the start as well as `00dcH264`. This is followed up by packets of varying size (mostly 1460 bytes) containing the camera stream, which is received until the connection is terminated

When setting your camera channel to one which does not have a camera connected via BNC, the stream still works. However, you may see `31dcH264` a lot and if you use the `nc` command, you may repeatedly hear the system bell sound (probably due to the terminal displaying the bell character).



