package main

import (
	"bufio"
	"fmt"
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"os"
	"errors"
	"strings"
	"time"
	"github.com/jogi1/poe"
)

type Config struct {
	AccountName string
	Email       string
	Password    string
	StashLeagues []string
	DelayTime time.Duration
}

func removeNameStuff(name string) string {
	splits := strings.Split(name, ">>")
	return splits[len(splits)-1]
}

func displayStartupError(message error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println(message)
	fmt.Println("Press enter to close this window.")
	reader.ReadString('\n')
}

func loadConfig() (config Config, err error) {
	var conf Config
	tomlData, err := ioutil.ReadFile("config")
	if err != nil {
		return conf, errors.New("could not open config file")
	}

	if _, err := toml.Decode(string(tomlData), &conf); err != nil {
		// handle error
		return conf, err
	}

	if len(conf.AccountName) == 0 {
		return conf, errors.New("need an AccountName")
	}

	if len(conf.Email) == 0 {
		return conf, errors.New("need an Email")
	}

	if len(conf.Password) == 0 {
		return conf, errors.New("need an Password")
	}
	return conf, nil
}


func printCharName (name string, league string) {
	s := fmt.Sprintf("%v in League %v", name, league)
	fmt.Printf("\r%v%*v\n", s, 60 - len(s), " ")
}

func poeFindCharacterMTX (p *poe.Poe, conf Config) {
	characters, err := p.GetCharacters()
	if err != nil {
		displayStartupError(err)
		return
	}

	for index, character := range characters {
		var charNamePrinted = bool(false)
		fmt.Printf("\r%60v", " ")
		s := fmt.Sprintf("Checking %2v/%2v (%v) in League (%v)", index +1, len(characters), character.Name, character.League)
		fmt.Printf("\r%v%*v", s, 60 - len(s), " ")
		charItems, err := p.GetCharacterItems(character.Name)
		if err != nil {
			displayStartupError(err)
			return
		}

		for _, item := range charItems.Items {
			if len(item.CosmeticMods) > 0 {
				if !charNamePrinted {
					charNamePrinted = true
					printCharName(character.Name, character.League)
				}
				fmt.Printf("\t %v (%v)-> %v\n", removeNameStuff(item.Name), item.TypeLine, item.CosmeticMods)
			}
			for _, socketedItem := range item.SocketedItems {
				if len(socketedItem.CosmeticMods) > 0 {
					if !charNamePrinted {
						charNamePrinted = true
						printCharName(character.Name, character.League)
					}
					fmt.Printf("\t %v (%v)-> %v socketed in (%v - %v)\n", removeNameStuff(socketedItem.Name), socketedItem.TypeLine, socketedItem.CosmeticMods, removeNameStuff(item.Name), item.TypeLine)
				}
			}
		}
	}
	fmt.Printf("\r%*v\n", 60, " ")
}


func printStashCount(league string, stashtab string, index int, max int) {
	s := fmt.Sprintf("%v %2v/%2v - %v", league, index + 1, max, stashtab)
	fmt.Printf("\r%v%*v", s, 60 - len(s), " ")
}

func printTab(league string, stash string) {
	s := fmt.Sprintf("%v - %v", league, stash)
	fmt.Printf("\r%v%*v\n", s, 60 - len(s), " ")
}

func printItem(item poe.Item) {
	fmt.Printf("\t%v (%v) (%v|%v) -> %v\n", removeNameStuff(item.Name), item.TypeLine, item.X + 1, item.Y + 1, item.CosmeticMods)
}

func printSocketedItem(item poe.Item, parent poe.Item) {
	fmt.Printf("\t%v (%v|%v) -> %v socketed in %v (%v) (%v|%v)\n", item.TypeLine, item.X + 1, item.Y + 1, item.CosmeticMods, removeNameStuff(parent.Name), parent.TypeLine, parent.X +1, parent.Y+1)
}

func poeFindStashMTX(p *poe.Poe, conf Config) {
	// getting stash
	for _, league := range conf.StashLeagues {
		var tabNamePrinted = bool(false)
		firstTab, err := p.GetStash(league, 0, 1)
		if err != nil {
			displayStartupError(err)
			return
		}

		if firstTab.NumTabs == 0 {
			continue
		}

		printStashCount(league, firstTab.Tabs[0].N, 0, firstTab.NumTabs)
		time.Sleep(1000 * time.Millisecond * conf.DelayTime)

		for _, item := range firstTab.Items {
			if len(item.CosmeticMods) > 0 {
				if !tabNamePrinted {
					tabNamePrinted = true
					printTab(league, firstTab.Tabs[0].N)
				}
			}
		}

		for i := 1; i < firstTab.NumTabs; i++ {
			currentTab, err := p.GetStash(league, i, 0)
			if err != nil {
				displayStartupError(err)
				return
			}
			printStashCount(league, firstTab.Tabs[i].N, i, firstTab.NumTabs)

			for _, item := range currentTab.Items {
				if len(item.CosmeticMods) > 0 {
					if !tabNamePrinted {
						tabNamePrinted = true
						printTab(league, firstTab.Tabs[i].N)
					}
					printItem(item)
				}
				for _, socketedItem := range item.SocketedItems {
					if len(socketedItem.CosmeticMods) > 0 {
						if !tabNamePrinted {
							tabNamePrinted = true
							printTab(league, firstTab.Tabs[i].N)
						}
						printSocketedItem(socketedItem, item)
					}
				}
			}

			time.Sleep(1000 * time.Millisecond * conf.DelayTime)
		}
	}
}

func main() {
	conf, err := loadConfig()
	if err != nil {
		displayStartupError(err)
		return
	}

	p := new(poe.Poe)
	err = p.Login(conf.AccountName, conf.Email, conf.Password)
	if err != nil {
		displayStartupError(err)
		return
	}

	poeFindCharacterMTX(p, conf)

	poeFindStashMTX(p, conf)

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Press enter to close this window.\n")
	reader.ReadString('\n')
}
