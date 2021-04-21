# Edit GMail filters

How to use:

- Go to the GMail filter page
  - Either
    - Visit https://mail.google.com/mail/u/0/#settings/filters
    - Or in GMail
      - Click the gear icon
      - Click `Settings`
      - Click `Filters`
  - At the bottom click Select: `All`
  - Click `Export`
- `go run github.com/maruel/gmailfilter mailFilters.xml > a.txt`
- Open https://sheets.new
  - Click File, Import
  - Click on tab `Upload`
  - Drag a.txt into the spreadsheet, give is a good 2 seconds, it's a bit slow.
  - Accept defaults


## Authors

`gmailfilter` was created with ❤️️ and passion by [Marc-Antoine
Ruel](https://github.com/maruel).


## Disclaimer

This is not an official Google product (experimental or otherwise), it
is just code that happens to be owned by Google.
