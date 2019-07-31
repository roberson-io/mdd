# MILLION DOLLAR DREAM (Golang edition)
Every man has a price.

This creates a Bloom filter to store and lookup file hashes in a space efficent
manner. 

[The original version](https://github.com/droberson/million_dollar_dream) of this tool was written in Python.

## Dependencies
[Go wrapper for MurmurHash3](https://github.com/roberson-io/mmh3)

## Usage
### Calculate hashes and store in a new Bloom filter file
```bash
./mdd calculate <filterfile> <directory>
```
For example:
```bash
./mdd calculate ./filters/wordpress /tmp/wordpress
```

### Lookup files in a directory using an existing filter
```bash
./mdd lookup <filterfile> <directory>
```

Maybe you generated the Wordpress filter using the calculate command above and want to check an installation of Wordpress:
```bash
./mdd lookup ./filters/wordpress /path/to/wordpress
```

### Create a new Bloom filter with a text file containing MD5 hashes
```bash
./mdd fromfile <filterfile> <hashfile>
```

For example, you have a `hashes.txt` file containing MD5 hashes for files in an application:
```
793e9490b89f2246eb644d70f4504140
4712e995ba48f00911e23ab6230808e2
ec0e6f5c28f6b251563f42adf6f47544
...
```

```bash
./mdd fromfile ./filters/myapp ./hashes.txt
```

Any lines that are not 32-character hex strings will be ignored.

### Fetch filter files from a remote repository
By default, this tool points to [my mdd_filters GitHub repository](https://github.com/roberson-io/mdd_filters/raw/master/repo/). The first time you run a `filters` command, the tool will create a `config.json` file.  You can edit `config.json` to point anywhere that serves a `METADATA.json` file and filter files from the same endpoint via HTTP.  There is a Python script in my mdd_filters repo that generates `METADATA.json`.

#### List locally installed filter files
```bash
./mdd filters list
```

#### List remote filter files at remote endpoint set in config.json
```bash
./mdd filters list remote
```

#### List remote filter files at some other URL
```bash
./mdd filters list <url>
```

#### Fetch a remote filter file
```bash
./mdd filters fetch <filter_name>
```

#### Pull any updated filter files from remote
```bash
./mdd filters update
```
