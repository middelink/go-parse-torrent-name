package parsetorrentname

import (
	"reflect"
	"strconv"
	"strings"
)

// TorrentInfo is the resulting structure returned by Parse
type TorrentInfo struct {
	Title      string
	Season     int    `json:"season,omitempty"`
	Episode    int    `json:"episode,omitempty"`
	Year       int    `json:"year,omitempty"`
	Resolution string `json:"resolution,omitempty"`
	Quality    string `json:"quality,omitempty"`
	Codec      string `json:"codec,omitempty"`
	Audio      string `json:"audio,omitempty"`
	Group      string `json:"group,omitempty"`
	Region     string `json:"region,omitempty"`
	Extended   bool   `json:"extended,omitempty"`
	Hardcoded  bool   `json:"hardcoded,omitempty"`
	Proper     bool   `json:"proper,omitempty"`
	Repack     bool   `json:"repack,omitempty"`
	Container  string `json:"container,omitempty"`
	Widescreen bool   `json:"widescreen,omitempty"`
	Website    string `json:"website,omitempty"`
	Language   string `json:"language,omitempty"`
	Sbs        string `json:"sbs,omitempty"`
	Unrated    bool   `json:"unrated,omitempty"`
	Size       string `json:"size,omitempty"`
	ThreeD     bool   `json:"3d,omitempty"`
}

func setField(tor *TorrentInfo, field, raw, val string) {
	ttor := reflect.TypeOf(tor)
	torV := reflect.ValueOf(tor)
	field = strings.Title(field)
	v, _ := ttor.Elem().FieldByName(field)
	//fmt.Printf("    field=%v, type=%+v, value=%v\n", field, v.Type, val)
	switch v.Type.Kind() {
	case reflect.Bool:
		torV.Elem().FieldByName(field).SetBool(true)
	case reflect.Int:
		clean, _ := strconv.ParseInt(val, 10, 64)
		torV.Elem().FieldByName(field).SetInt(clean)
	case reflect.Uint:
		clean, _ := strconv.ParseUint(val, 10, 64)
		torV.Elem().FieldByName(field).SetUint(clean)
	case reflect.String:
		torV.Elem().FieldByName(field).SetString(val)
	}
}

// Parse breaks up the given filename in TorrentInfo
func Parse(filename string) (*TorrentInfo, error) {
	tor := &TorrentInfo{}
	//fmt.Printf("filename %q\n", filename)

	var startIndex, endIndex = 0, len(filename)
	for _, pattern := range patterns {
		cleanName := strings.Replace(filename, "_", " ", -1)
		matches := pattern.re.FindStringSubmatch(cleanName)
		if len(matches) == 0 {
			continue
		}
		//fmt.Printf("  %s: pattern:%q match:%#v\n", pattern.name, pattern.re, matches)

		index := strings.Index(filename, matches[1])
		if index == 0 {
			startIndex = len(matches[1])
			//fmt.Printf("    startIndex moved to %d [%q]\n", startIndex, filename[startIndex:endIndex])
		} else if index < endIndex {
			endIndex = index
			//fmt.Printf("    endIndex moved to %d [%q]\n", endIndex, filename[startIndex:endIndex])
		}
		setField(tor, pattern.name, matches[1], matches[2])
	}

	// Start process for title
	//fmt.Println("  title: <internal>")
	raw := strings.Split(filename[startIndex:endIndex], "(")[0]
	clean := raw
	if strings.HasPrefix(clean, "- ") {
		clean = raw[2:]
	}
	if strings.ContainsRune(clean, '.') && !strings.ContainsRune(clean, ' ') {
		clean = strings.Replace(clean, ".", " ", -1)
	}
	clean = strings.Replace(clean, "_", " ", -1)
	//clean = re.sub('([\[\(_]|- )$', '', clean).strip()
	setField(tor, "title", raw, strings.TrimSpace(clean))

	return tor, nil
}

/*
class PTN(object):
    def _escape_regex(self, string):
        return re.sub('[\-\[\]{}()*+?.,\\\^$|#\s]', '\\$&', string)

    def __init__(self):
        self.parts = None
        self.torrent = None
        self.excess_raw = None
        self.group_raw = None
        self.start = None
        self.end = None

    def _part(self, name, match, raw, clean):
        # The main core instructuions
        self.parts[name] = clean

        if len(match) != 0:
            # The instructions for extracting title
            index = self.torrent['name'].find(match[0])
            if index == 0:
                self.start = len(match[0])
            elif self.end is None or index < self.end:
                self.end = index

        if name != 'excess':
            # The instructions for adding excess
            if name == 'group':
                self.group_raw = raw
            if raw is not None:
                self.excess_raw = self.excess_raw.replace(raw, '')

    def _late(self, name, clean):
        if name == 'group':
            self._part(name, [], None, clean)
        elif name == 'episodeName':
            clean = re.sub('[\._]', ' ', clean)
            clean = re.sub('_+$', '', clean)
            self._part(name, [], None, clean.strip())

    def parse(self, name):
        self.parts = {}
        self.torrent = {'name': name}
        self.excess_raw = name
        self.group_raw = ''
        self.start = 0
        self.end = None

        for key, pattern in patterns:
            if key not in ('season', 'episode', 'website'):
                pattern = r'\b%s\b' % pattern

            clean_name = re.sub('_', ' ', self.torrent['name'])
            match = re.findall(pattern, clean_name, re.I)
            if len(match) == 0:
                continue
	    #print "%s: %s -> %s" % (name, pattern, match)

            index = {}
            if isinstance(match[0], tuple):
                match = list(match[0])
            if len(match) > 1:
                index['raw'] = 0
                index['clean'] = 1
            else:
                index['raw'] = 0
                index['clean'] = 0

            if key in types.keys() and types[key] == 'boolean':
                clean = True
            else:
                clean = match[index['clean']]
                if key in types.keys() and types[key] == 'integer':
                    clean = int(clean)
            if key == 'group':
                if re.search(patterns[5][1], clean, re.I) \
                        or re.search(patterns[4][1], clean):
                    continue  # Codec and quality.
                if re.match('[^ ]+ [^ ]+ .+', clean):
                    key = 'episodeName'
            if key == 'episode':
                sub_pattern = self._escape_regex(match[index['raw']])
                self.torrent['map'] = re.sub(
                    sub_pattern, '{episode}', self.torrent['name']
                )
            self._part(key, match, match[index['raw']], clean)

        # Start process for title
        raw = self.torrent['name']
        if self.end is not None:
            raw = raw[self.start:self.end].split('(')[0]

        #print "raw=", raw
        clean = re.sub('^ -', '', raw)
        if clean.find(' ') == -1 and clean.find('.') != -1:
            clean = re.sub('\.', ' ', clean)
        clean = re.sub('_', ' ', clean)
        clean = re.sub('([\[\(_]|- )$', '', clean).strip()
	#print "clean=", clean

        self._part('title', [], raw, clean)

        # Start process for end
        clean = re.sub('(^[-\. ()]+)|([-\. ]+$)', '', self.excess_raw)
        clean = re.sub('[\(\)\/]', ' ', clean)
        match = re.split('\.\.+| +', clean)
        if len(match) > 0 and isinstance(match[0], tuple):
            match = list(match[0])

        clean = filter(bool, match)
        clean = [item for item in filter(lambda a: a != '-', clean)]
        clean = [item.strip('-') for item in clean]
        if len(clean) != 0:
            group_pattern = clean[-1] + self.group_raw
            if self.torrent['name'].find(group_pattern) == \
                    len(self.torrent['name']) - len(group_pattern):
                self._late('group', clean.pop() + self.group_raw)

            if 'map' in self.torrent.keys() and len(clean) != 0:
                episode_name_pattern = (
                    '{episode}'
                    '' + re.sub('_+$', '', clean[0])
                )
                if self.torrent['map'].find(episode_name_pattern) != -1:
                    self._late('episodeName', clean.pop(0))

        if len(clean) != 0:
            if len(clean) == 1:
                clean = clean[0]
            self._part('excess', [], self.excess_raw, clean)
        return self.parts
*/
