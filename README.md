# Little Explorer

A server-to-server web app to consume NASA API's portal. APOD (Astronomy Picture Of the Day) API consuming is now implemented. More API's will follow.

## Installation.
`cd` to `$GOPATH/src/github` directory of your Go installation and clone with the following command:
```
$ git clone https://github.com/niconc/littleExplorer.git
```

## Usage.
**Step #1.**
Open `terminal` or `iterm` or any terminal emulator, and run the server with the following command:
```
$ go run littleExplorer.go

```
**Step #2.**
You will be prompted to enter the [NASA's API Key.](https://api.nasa.gov/){:target="_blank"} Enter your API Key from [NASA's portal](https://api.nasa.gov/){:target="_blank"} (if you have one) otherwise `DEMO_KEY` will be used.

**Step #3.**
Open a browser and run the server locally by using `localhost:3000/apod/2020-12-25`. This will call the NASA's APOD API and return the picture of the date given at the address bar (2020-12-25). You may also not entering any date at all, like `localhost:3000/apod`, and in that case, the current day's picture of the day will be returned.

If you like Astronomy, Enjoy!!!

_****BTW, Merry Christmas & a Happy New Year for 2021!***
