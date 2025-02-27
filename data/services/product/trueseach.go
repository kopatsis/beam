package product

import (
	"math"
	"regexp"
	"sort"
	"strings"

	"beam/data/models"

	"github.com/texttheater/golang-levenshtein/levenshtein"
)

type Searcher struct {
	ProductInfo     models.ProductInfo
	InitialScore    int // Out of 1000
	TitleMatchScore int // Out of 100
	SalesScore      int // Out of 100
	FinalScore      int
}

func (s *Searcher) TallyScore() {
	s.FinalScore = s.InitialScore + 3*s.SalesScore + 4*s.TitleMatchScore
}

func cleanSearchString(input string) string {
	if len(input) > 140 {
		input = input[:140]
	}

	input = strings.ToLower(input)

	re := regexp.MustCompile("[^a-zA-Z0-9 ]+")
	input = re.ReplaceAllString(input, "")

	reSpace := regexp.MustCompile(`\s+`)
	input = reSpace.ReplaceAllString(input, " ")

	return input
}

func fuzzyCompare(target, str string, strict bool) int {
	if target == "" || str == "" {
		return 0
	}

	if target == str {
		return 25
	}

	customOptions := levenshtein.Options{
		InsCost: 2,
		DelCost: 2,
		SubCost: 3,
		Matches: levenshtein.DefaultOptions.Matches,
	}

	distance := levenshtein.DistanceForStrings([]rune(target), []rune(str), customOptions)
	var score int

	if strict {
		score = 15 - (distance / 2)
	} else {
		score = 20 - (distance / 2)
	}

	if score < 0 {
		return 0
	}
	return score
}

func scoreSales(prods []*Searcher) {
	max := 0
	for _, p := range prods {
		if p.ProductInfo.Sales > max {
			max = p.ProductInfo.Sales
		}
	}

	if max <= 0 {
		return
	}

	for i := range prods {
		prods[i].SalesScore = int(math.Round(100 * float64(prods[i].ProductInfo.Sales) / float64(max)))
	}
}

func filterSearchers(searchers []*Searcher) []*Searcher {
	var filtered []*Searcher
	for _, s := range searchers {
		if s != nil && s.InitialScore > 100 {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

func scorePortions(tokens []string, prods []*Searcher) {
	scoreTitling(tokens, prods)
	scoreSKUs(tokens, prods)
	scoreRegTags(tokens, prods)
	scoreFilterTags(tokens, prods)
	scoreVariants(tokens, prods)
}

func tokenPositionAndFactor(score, position, factor int) int {
	score *= factor
	if position == 0 {
		score *= 6
	} else if position == 1 {
		score *= 4
	} else if position == 3 || position == 4 {
		score *= 3
	} else {
		score *= 2
	}
	return score
}

func tokenLengthAndFactor(score, length int) int {
	if length == 1 {
		score /= 6
	} else if length == 2 {
		score /= 10
	} else if length == 3 {
		score /= 13
	} else if length == 4 {
		score /= 16
	} else {
		score /= 18
	}
	return score
}

func scoreSKUs(tokens []string, prods []*Searcher) {
	factor := 7
	for i, prod := range prods {
		skuScore := 0
		for _, sku := range prod.ProductInfo.SKUs {
			currentScore := 0

			for i, token := range tokens {
				currentScore += tokenPositionAndFactor(fuzzyCompare(token, sku, true), i, factor)
			}

			skuScore += tokenLengthAndFactor(currentScore, len(tokens))
		}
		skuScore /= len(prod.ProductInfo.SKUs)
		prods[i].InitialScore += skuScore
	}
}

func scoreVariants(tokens []string, prods []*Searcher) {
	factor := 6
	for i, prod := range prods {
		varScore := 0
		variants := append(prod.ProductInfo.Var1Values, prod.ProductInfo.Var2Values...)
		variants = append(variants, prod.ProductInfo.Var3Values...)

		if len(variants) == 0 || len(variants) == 1 && variants[0] == "*" {
			return
		}

		for _, variant := range variants {
			currentScore := 0

			for i, token := range tokens {
				currentScore += tokenPositionAndFactor(fuzzyCompare(token, variant, false), i, factor)
			}

			varScore += tokenLengthAndFactor(currentScore, len(tokens))
		}

		varScore /= len(variants)
		prods[i].InitialScore += varScore
	}
}

func scoreRegTags(tokens []string, prods []*Searcher) {
	factor := 9
	for i, prod := range prods {
		tagScore := 0
		count := 0

		for _, tag := range prod.ProductInfo.SKUs {
			tag = strings.TrimSpace(tag)
			if strings.Contains(tag, "__") || tag == "" {
				continue
			}

			currentScore := 0

			for i, token := range tokens {
				currentScore += tokenPositionAndFactor(fuzzyCompare(token, tag, true), i, factor)
			}

			tagScore += tokenLengthAndFactor(currentScore, len(tokens))
			count++
		}

		if count > 0 {
			tagScore /= count
		} else {
			tagScore = 0
		}
		prods[i].InitialScore += tagScore
	}
}

func scoreFilterTags(tokens []string, prods []*Searcher) {
	factor := 4
	for i, prod := range prods {
		tagScore := 0
		count := 0

		for _, tag := range prod.ProductInfo.SKUs {
			tag = strings.TrimSpace(tag)
			if !strings.Contains(tag, "__") {
				continue
			}

			if index := strings.Index(tag, "__"); index != -1 {
				tag = tag[index+2:]
			} else {
				tag = ""
			}
			if tag == "" {
				continue
			}

			currentScore := 0

			for i, token := range tokens {
				currentScore += tokenPositionAndFactor(fuzzyCompare(token, tag, true), i, factor)
			}

			tagScore += tokenLengthAndFactor(currentScore, len(tokens))
			count++
		}

		if count > 0 {
			tagScore /= count
		} else {
			tagScore = 0
		}
		prods[i].InitialScore += tagScore
	}
}

func scoreTitling(tokens []string, prods []*Searcher) {
	factor := 14
	for i, prod := range prods {
		titleScore := 0
		eachInTitle := strings.Split(prod.ProductInfo.Title, " ")
		count := 0

		for _, word := range eachInTitle {
			word = strings.TrimSpace(word)
			if word == "" {
				continue
			}

			currentScore := 0

			for i, token := range tokens {
				currentScore += tokenPositionAndFactor(fuzzyCompare(token, word, false), i, factor)
			}

			titleScore += tokenLengthAndFactor(currentScore, len(tokens))
			count++
		}

		if count > 0 {
			titleScore /= count
		} else {
			titleScore = 0
		}
		prods[i].InitialScore += titleScore
	}
}

func RemoveNonMatches(prods []*Searcher) []*Searcher {
	var filteredProds []*Searcher
	for _, prod := range prods {
		if prod.InitialScore >= 15 || prod.TitleMatchScore > 50 {
			filteredProds = append(filteredProds, prod)
		}
	}
	return filteredProds
}

func ScoreFullTitleComp(full string, prods []*Searcher) {
	for i := range prods {
		prods[i].TitleMatchScore = 4 * fuzzyCompare(full, prods[i].ProductInfo.Title, true)
	}
}

func splitTokens(initial string) (string, []string) {
	initial = cleanSearchString(initial)

	splits := strings.Split(initial, " ")

	ret := []string{}
	for _, token := range splits {
		if len(ret) >= 5 {
			break
		}
		token = strings.TrimSpace(token)
		if len(token) > 1 || token == "u" || token == "i" || token == "a" {
			ret = append(ret, token)
		}
	}

	if len(ret) == 0 {
		return "", ret
	}

	return initial, ret
}

func FullSearch(initial string, provided []models.ProductInfo) []models.ProductInfo {

	searchers := []*Searcher{}
	for _, p := range provided {
		searchers = append(searchers, &Searcher{
			ProductInfo: p,
		})
	}
	scoreSales(searchers)

	full, tokens := splitTokens(initial)
	if full != "" {
		scorePortions(tokens, searchers)
		searchers = filterSearchers(searchers)
	}

	for i, searcher := range searchers {
		searcher.TallyScore()
		searchers[i] = searcher
	}

	sort.SliceStable(searchers, func(i, j int) bool {
		if searchers[i].FinalScore != searchers[j].FinalScore {
			return searchers[i].FinalScore > searchers[j].FinalScore
		}
		if searchers[i].SalesScore != searchers[j].SalesScore {
			return searchers[i].SalesScore > searchers[j].SalesScore
		}
		return searchers[i].ProductInfo.Title < searchers[j].ProductInfo.Title
	})

	ret := []models.ProductInfo{}
	for _, searcher := range searchers {
		ret = append(ret, searcher.ProductInfo)
	}

	return ret
}
