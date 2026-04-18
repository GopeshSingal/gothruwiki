# gothruwiki

A tool for streaming official releases of Wikipedia's XML ``.bz2`` file datasets and to run custom processing / analyses scripts.

## Background

For a personal project, I wanted to determine the most common words on Wikipedia. It was through this endeavor that I discovered [WikiExtractor](https://github.com/attardi/wikiextractor). While I extend enormous respect towards its contributors, I found myself frustrated with using the tool due to three primary reasons:

1. The tool itself appeared to be unmaintained and deprecated
2. The runtime of the tool was not fantastic by itself
3. It was rather single-focused and was not open towards modularity

From this experience, I decided to try making my own means of processing the Wikipedia dataset dump, deciding to use Go (not much reasearch into this decision, mostly on a whim). 

## Purpose and Scope

The purpose of ``gothruwiki`` is to go through Wikipedia's data dump and allow a user to use streaming for processing or analyzing the data for whatever purpose they require. 

This is NOT a general XML tool. For such, please refer to the package being used for the fast XML parsing (and give them a star): [helium](https://github.com/lestrrat-go/helium). While advertised as for Wikipedia, there's a high chance you could use this for other XML datasets if need be. 

## Usage

### Recommendation

For best performance, it is highly encouraged to install [lbzip2](https://github.com/kjn/lbzip2), which reduces the runtime of this process immensely. Please star them while you're at it. 

### Requirements

Match the ``go.mod`` file

Download [enwiki-latest-pages-articles.xml.bz2](https://dumps.wikimedia.org/enwiki/latest/enwiki-latest-pages-articles.xml.bz2)

### Build

To build the binary simply run the following from the root directory: ``go build -o gothruwiki .``

### Configuration

The file ``gothruwiki.json.example`` outlines the available flags that can be utilized. You can create your own ``.json`` file to store your configurations for running, or use the flags manually. 

### Processors

The repo includes two processors: ``wordcount`` and ``linkgraph`` to serve basic needs and to be examples for custom processors. For implementing your own processor, please refer to ``processors/CUSTOM.md``. Once that has been implemented, update ``registry.go`` to link to your custom-built processor. 

If you think your did a good job, then you may as well make a pull request and publish your processor for the world to see. Why not, right?

### Running

``./gothruwiki -config <config.json>``