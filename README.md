# swanntools
Capture tools for the Swann DVR4-1200 DVR, inspired by [meatheadmike/swanntools](https://github.com/meatheadmike/swanntools) which was developed for DVR8-2600. Whereas the latter's mobile stream script is compatible with the DVR4-1200, the media script (featuring much higher quality) does not.

This repository is part of a long-term project to build a more secure system piggybacking off the DVR4-1200 with cloud backups and much better web/mobile clients.

## Roadmap

Current task: Conduct research into authentication using Wireshark

- [ ] Port existing code to Go (just to learn a new language and to use [koding/kite](https://github.com/koding/kite) later on)
- [ ] Create a Go script which can authenticate with the DVR via its media protocol
- [ ] Add streaming of cameras to the script
- [ ] Plan a method to stream the H264 stream to AWS/Azure in order to transcode, store and stream video

## Research

The following observations can be reproduced using Wireshark captures with the capture filter `host [DVR IP] and port [MEDIA PORT (9000)]` while using the web client (default port 85). All communication with the DVR is made over TCP.

A hex editor such as [hexed.it](hexed.it) will be very useful for interpreting the hexdumps below. In the case of hexed.it, remember to substitute the dummy characters (e.g., `X`) for real hex values and when pasting the data, specify that the data should be interpreted as hexadecimal.

You will also find that an ASCII to Hex converter may be useful. I recommend [asciitohex.com](http://www.asciitohex.com/).

#### Authentication

To authenticate, the steps are as follows:

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
