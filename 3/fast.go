package noymain

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

type  User struct{
	Browsers  []string `json:"browsers"`
	Name string        `json:"name"`
	Email string       `json:"email"`
}


//вам надо написать более быструю оптимальную этой функции
func FastSearch(out io.Writer) {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	fileContents, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}

	foundUsers := make([]string, 0)
	seenBrowsers := []string{}
	uniqueBrowsers := 0

	lines := strings.Split(string(fileContents), "\n")

	users := make([]User, 0, len(lines))
	for _, line := range lines {
		user := User{}
		// fmt.Printf("%v %v\n", err, line)
		err := user.UnmarshalJSON([]byte(line))
		if err != nil {
			panic(err)
		}
		users = append(users, user)
	}

	for i, user := range users {

		isAndroid := false
		isMSIE := false

		for _, browser := range user.Browsers {

			if ok := strings.Contains(browser, "Android"); ok {
				isAndroid = true
				notSeenBefore := true
				for _, item := range seenBrowsers {
					if item == browser {
						notSeenBefore = false
					}
				}
				if notSeenBefore {
					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
					seenBrowsers = append(seenBrowsers, browser)
					uniqueBrowsers++
				}
			}

			if ok := strings.Contains(browser, "MSIE"); ok {
				isMSIE = true
				notSeenBefore := true
				for _, item := range seenBrowsers {
					if item == browser {
						notSeenBefore = false
					}
				}
				if notSeenBefore {
					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
					seenBrowsers = append(seenBrowsers, browser)
					uniqueBrowsers++
				}
			}
		}


		if !(isAndroid && isMSIE) {
			continue
		}

		// log.Println("Android and MSIE user:", user["name"], user["email"])
		//foundUsers += fmt.Sprintf("[%d] %s <%s>\n", i, user.Name, email)
		email := strings.ReplaceAll(user.Email, "@"," [at] ")
		foundUsers = append(foundUsers, fmt.Sprintf("[%d] %s <%s>\n", i, user.Name, email))
	}


	fmt.Fprintln(out, "found users:\n"+ strings.Join(foundUsers, ""))
	fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))
}