# HTTP Server for Rødecaster Duo

This HTTP server runs on the Rødecaster Duo. It makes the Rødecaster features available over HTTP.

Right now, only the buttons can be pressed.

# Compiling

Type `make`.

# Installation

1. Connect the Rødecaster to Wi-Fi or cabled LAN.
2. Find out the IP address of the Rødecaster by tapping ⚙ -> System -> Network -> Advanced on the screen.
3. Run:
```sh
<rc2http ssh root@192.168.x.y \
  'cat>/tmp/rc2http;chmod a+x /tmp/rc2http;/tmp/rc2http --install-service;/etc/init.d/rc2http start'
```

replacing `192.168.x.y` with your IP address from step 2.
The password is `Yojcakhev90` .

4. To use the server, go to `http://192.168.x.y` in your web browser , again replacing `192.168.x.y` with your Rødecaster's IP address.

