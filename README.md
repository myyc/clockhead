clockhead
=========

A CPU frequency scaling daemon for Linux with no configuration.

`# sudo go run *.go`

It outputs whatever it's doing so you can watch the magic on
your terminal for a few seconds until you get bored. It even
has emojis!!!

The algorithm
=============

It's very sophisticated. It checks core load. If it's low it
goes down. If it's very low it goes even more down. If it's high
it goes up. If it's very high, you get the deal.

Oh, and if your laptop is plugged to AC it sets the governor to
`performance` and doesn't do anything.

If you want to disable it temporarily while keeping it running
create `/tmp/clockhead.lock`.

That's it. This is the pinnacle of my Linux Power Management
knowledge so if you want more features it's very likely I have
no idea how to implement them.
