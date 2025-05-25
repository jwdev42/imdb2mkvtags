# Documentation for imdb2mkvtags

## Synopsis

Generate the tags for the movie *Trading Places* and write the output to stdout: `imdb2mkvtags https://www.imdb.com/title/tt0086465`

Generate the tags for the movie *Trading Places*, also include full credits and all keywords. Write output to file *tags.xml*: `imdb2mkvtags -o tags.xml -opts fullcredits=1:keywords=1 https://www.imdb.com/title/tt0086465`

## Description

imdb2mkvtags scrapes information off the [internet movie database](<https://www.imdb.com/>) and writes it as a xml file containing matroska tags. This xml file can be processed by [MKVToolNix](<https://mkvtoolnix.download/>)	to tag mkv files.

## IMDB scraper module

The IMDB scraper module will be used on the following input URLs:

- `{"http"|"https"}://www.imdb.com/title/{MOVIEID}`
- `{"http"|"https"}://www.imdb.com/{LANGUAGE}/title/{MOVIEID}`
- `imdb://{MOVIEID}`

| Token    | Description |
| -------- | ------- |
| MOVIEID  | The IMDB movie ID, a string that starts with `tt` and is followed by an integer. Example: `tt1136608` for the movie *District 9*. |
| LANGUAGE | ISO 639-1 language code. Examples: `en` for *English*, `de` for *German*. |

### IMDB scraper options

#### \-lang *language*

Specifies the preferred language you want to receive the content in. The token *language* must obey the format	*xx-XX*, where *xx* must be substituted with an ISO 639-1 code (small letters) and *XX* must be substituted	with an Alpha-2 code (capital letters).

#### \-loglevel *loglevel*

Specifies the least significant loglevel to be displayed. Available loglevels are:

- panic
- alert
- critical
- error
- warning
- notice
- info
- debug

Default value is *notice*. It is not recommended to set a value higher than *error*.

#### \-opts *options*

Specifies the options specific to the website that is used for scraping. The option string must conform to the following spec: `option1=value1:option2=value2:optionN=valueN`  
Values of type *bool* can be *true* | *false* | *1* | *0*.

##### Options for imdb.com

###### keywords=*bool*

Additionally scrapes IMDB's keyword page for the given movie if enabled. Disabled by default.

###### keyword-limit=*int*

Limits the keywords for the tag file to accept to the specified amount. Must be a positive integer > 0 to be enabled. Default value is 0.
Only in effect if option keywords is *true*. Set it to 5 to mimic the behaviour of the title page.

###### fullcredits=*bool*

Additionally scrapes IMDB's fullcredits page for the given movie if enabled. Disabled by default. Only scrapes the 50 most relevant tags as IMDB does not make the full dataset available in the html document anymore.

###### jsonld=*bool*

If enabled, the data source for the title page information will be the embedded json-ld data instead of the title page itself.	Can be used as a backup if the normal title page scraper fails. Disabled by default.

### IMDB scraper issues and limitations

If you use the option `keywords=1`, only the first 50 keywords will be scraped. Scraping all keywords would require javascript interpretation and will therefore never be supported.

IMDB frequently update their website to make it harder to scrape. It might be that this scraper will stop working sometimes in the future.

### IMDB scraper troubleshooting

The IMDB scraper often breaks after IMDB website updates. You might try the option `jsonld=1` to work around this. Feel free to file an issue if you encountered such a problem.
