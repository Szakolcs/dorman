package main

import (
	"dorm-man/internal/administration"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

type inputData struct {
	Attributes administration.TenantMassAssignmentRequest `json:"request"`
	Tenants    []administration.MassAssignmentTenant      `json:"tenants"`
	Flats      []administration.MassAssignmentFlat        `json:"flats"`
}

type strictGroup struct {
	Name    string                                `json:"name"`
	Tenants []administration.MassAssignmentTenant `json:"tenants"`
	Flats   []administration.MassAssignmentFlat   `json:"flats"`
}

func (s *strictGroup) AddTenant(tenant administration.MassAssignmentTenant) {
	s.Tenants = append(s.Tenants, tenant)
}

func (s *strictGroup) AddFlat(flat administration.MassAssignmentFlat) {
	s.Flats = append(s.Flats, flat)
}
func GroupUsers(users []administration.MassAssignmentTenant, attributes []string) (map[string][]administration.MassAssignmentTenant, error) {
	if len(users) == 0 || len(attributes) == 0 {
		return make(map[string][]administration.MassAssignmentTenant), nil
	}

	grouped := make(map[string][]administration.MassAssignmentTenant)

	for _, u := range users {
		rv := reflect.ValueOf(u)
		var keyParts []string

		for _, attr := range attributes {
			fieldName := strings.ToUpper(string(attr[0])) + attr[1:]
			field := rv.FieldByName(fieldName)
			if !field.IsValid() {
				return nil, fmt.Errorf("invalid attribute %q: not found in User struct", attr)
			}
			if field.Kind() == reflect.String {
				keyParts = append(keyParts, field.String())
			} else if field.Kind() == reflect.Int {
				keyParts = append(keyParts, strconv.Itoa(int(field.Int())))
			}
		}

		key := strings.Join(keyParts, "|||")
		grouped[key] = append(grouped[key], u)
	}

	return grouped, nil
}

func main() {
	data, err := os.ReadFile("./cmd/test.json")
	if err != nil {
		panic(err)
	}
	var inputData inputData
	err = json.Unmarshal(data, &inputData)
	if err != nil {
		panic(err)
	}
	for i := range inputData.Flats {
		for _, room := range inputData.Flats[i].Rooms {
			inputData.Flats[i].Capacity += room.Capacity
		}
	}
	strictGroups := strings.Split(inputData.Attributes.StrictGroups, ",")
	groupedTenants, err := GroupUsers(inputData.Tenants, strictGroups)
	flatArray := make([]int, 0)
	for _, flat := range inputData.Flats {
		flatArray = append(flatArray, flat.Capacity)
	}
	coinChanges := make(map[string][][]int)
	for key, group := range groupedTenants {
		sum := len(group)
		coins := flatArray
		coinChanges[key] = append(coinChanges[key], calculateCoinChange(coins, sum)...)
	}
	validArrangement, _ := FindValidArrangement(flatArray, coinChanges)
	strictGroups = make([]string, 0)
	for k, _ := range groupedTenants {
		strictGroups = append(strictGroups, k)
	}
	for _, group := range strictGroups {
		preferences := strings.Split(inputData.Attributes.Preferences, ",")
		preferenceGroups, err := GroupUsers(groupedTenants[group], preferences)
		if err != nil {
			panic(err)
		}
		preferences = make([]string, 0)
		for k, _ := range preferenceGroups {
			preferences = append(preferences, k)
		}
		sort.Strings(preferences)
		fmt.Println(validArrangement)
		fmt.Println(preferences)
	}
}

func calculateCoinChange(coins []int, amount int) [][]int {
	if amount < 0 || len(coins) == 0 {
		return [][]int{}
	}
	sort.Ints(coins)
	var result [][]int
	var path []int

	deptFirst(0, amount, path, &result, coins)
	return result
}

func deptFirst(startIdx int, remaining int, path []int, result *[][]int, coins []int) {
	if remaining == 0 {
		combo := make([]int, len(path))
		copy(combo, path)
		*result = append(*result, combo)
		return
	}

	for i := startIdx; i < len(coins); i++ {
		if i > startIdx && coins[i] == coins[i-1] {
			continue
		}
		if coins[i] > remaining {
			break
		}
		path = append(path, coins[i])
		deptFirst(i+1, remaining-coins[i], path, result, coins)
		path = path[:len(path)-1]
	}
}

type category struct {
	key  string
	opts [][]int
}

func FindValidArrangement(coins []int, subsets map[string][][]int) (map[string][]int, bool) {
	if len(subsets) == 0 {
		return nil, false
	}

	origFreq := make(map[int]int)
	for _, c := range coins {
		origFreq[c]++
	}

	cats := make([]category, 0, len(subsets))
	for k, v := range subsets {
		cats = append(cats, category{k, v})
	}
	sort.Slice(cats, func(i, j int) bool {
		return len(cats[i].opts) < len(cats[j].opts)
	})

	result := make(map[string][]int)

	currFreq := make(map[int]int, len(origFreq))
	for k, v := range origFreq {
		currFreq[k] = v
	}
	if arrangementSolver(0, currFreq, cats, &result) {
		return result, true
	}
	return nil, false
}

func arrangementSolver(idx int, freq map[int]int, cats []category, result *map[string][]int) bool {
	if idx == len(cats) {
		return true
	}
	k := cats[idx].key
	for _, candidate := range cats[idx].opts {
		var used []int
		possible := true
		for _, coin := range candidate {
			if freq[coin] > 0 {
				freq[coin]--
				used = append(used, coin)
			} else {
				possible = false
				break
			}
		}

		if possible {
			(*result)[k] = candidate
			if arrangementSolver(idx+1, freq, cats, result) {
				return true
			}
			delete(*result, k)
		}

		for _, coin := range used {
			freq[coin]++
		}
	}
	return false
}
