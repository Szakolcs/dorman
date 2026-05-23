package main

import (
	"dorm-man/internal/administration"
	"encoding/json"
	"fmt"
	"math"
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

type roomAssignment struct {
	FlatID   string
	RoomID   string
	Capacity int
	Tenants  []administration.MassAssignmentTenant
}

type roomSlot struct {
	FlatID   string
	RoomID   string
	Capacity int
	Tenants  []administration.MassAssignmentTenant
}

const greedyFallbackThreshold = 12

func GroupUsers(users []administration.MassAssignmentTenant, attributes []string) (map[string][]administration.MassAssignmentTenant, error) {
	if len(users) == 0 || len(attributes) == 0 {
		return make(map[string][]administration.MassAssignmentTenant), nil
	}

	grouped := make(map[string][]administration.MassAssignmentTenant)

	for _, u := range users {
		key, err := tenantPreferenceKey(u, attributes)
		if err != nil {
			return nil, err
		}
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
	if err != nil {
		panic(err)
	}
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
	validArrangement, ok := FindValidArrangement(flatArray, coinChanges)
	if !ok {
		panic("no valid flat arrangement found for strict groups")
	}
	strictGroupKeys := make([]string, 0, len(groupedTenants))
	for k := range groupedTenants {
		strictGroupKeys = append(strictGroupKeys, k)
	}
	sort.Strings(strictGroupKeys)
	validArrangementToFlats := make(map[string][]administration.MassAssignmentFlat)
	for s, sizes := range validArrangement {
		for _, size := range sizes {
			for i, flat := range inputData.Flats {
				if flat.Capacity == size {
					validArrangementToFlats[s] = append(validArrangementToFlats[s], flat)
					inputData.Flats = append(inputData.Flats[:i], inputData.Flats[i+1:]...)
					break
				}
			}
		}
	}
	preferenceAttrs := strings.Split(inputData.Attributes.Preferences, ",")
	for _, group := range strictGroupKeys {
		tenants := groupedTenants[group]
		flats := validArrangementToFlats[group]
		roomCapTotal := 0
		for _, flat := range flats {
			for _, room := range flat.Rooms {
				roomCapTotal += room.Capacity
			}
		}
		if roomCapTotal != len(tenants) {
			panic(fmt.Sprintf("strict group %q: room capacity %d does not match tenant count %d", group, roomCapTotal, len(tenants)))
		}
		assignments, score := optimizeTenants(tenants, flats, preferenceAttrs)
		fmt.Printf("strict group: %s (score: %d)\n", group, score)
		flatScores := make(map[string]int)
		for _, assignment := range assignments {
			roomScoreVal := roomScore(assignment.Tenants, preferenceAttrs)
			flatScores[assignment.FlatID] += roomScoreVal
			tenantIDs := make([]string, len(assignment.Tenants))
			for i, tenant := range assignment.Tenants {
				tenantIDs[i] = tenant.ID
			}
			fmt.Printf("  flat=%s room=%s tenants=%v room_score=%d\n", assignment.FlatID, assignment.RoomID, tenantIDs, roomScoreVal)
		}
		for i := range flats {
			flats[i].CompatibilityScore = flatScores[flats[i].ID]
			fmt.Printf("  flat=%s compatibility_score=%d\n", flats[i].ID, flats[i].CompatibilityScore)
		}
		fmt.Println()
	}
}

func tenantPreferenceKey(tenant administration.MassAssignmentTenant, attributes []string) (string, error) {
	rv := reflect.ValueOf(tenant)
	keyParts := make([]string, 0, len(attributes))
	for _, attr := range attributes {
		fieldName := strings.ToUpper(string(attr[0])) + attr[1:]
		field := rv.FieldByName(fieldName)
		if !field.IsValid() {
			return "", fmt.Errorf("invalid attribute %q: not found in User struct", attr)
		}
		if field.Kind() == reflect.String {
			keyParts = append(keyParts, field.String())
		} else if field.Kind() == reflect.Int {
			keyParts = append(keyParts, strconv.Itoa(int(field.Int())))
		}
	}
	return strings.Join(keyParts, "|||"), nil
}

func calculateCompatibility(attr1 string, attr2 string) int {
	score := 0
	attr1Collection := strings.Split(attr1, "|||")
	attr2Collection := strings.Split(attr2, "|||")
	for i := range attr1Collection {
		if attr1Collection[i] != attr2Collection[i] {
			score -= (len(attr1Collection) - i) * 5
		}
	}
	return score
}

func roomScore(tenants []administration.MassAssignmentTenant, preferenceAttrs []string) int {
	if len(tenants) < 2 {
		return 0
	}
	keys := make([]string, len(tenants))
	for i, tenant := range tenants {
		key, err := tenantPreferenceKey(tenant, preferenceAttrs)
		if err != nil {
			panic(err)
		}
		keys[i] = key
	}
	score := 0
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			score += calculateCompatibility(keys[i], keys[j])
		}
	}
	return score
}

func flattenRoomSlots(flats []administration.MassAssignmentFlat) []roomSlot {
	slots := make([]roomSlot, 0)
	for _, flat := range flats {
		for _, room := range flat.Rooms {
			slots = append(slots, roomSlot{
				FlatID:   flat.ID,
				RoomID:   room.ID,
				Capacity: room.Capacity,
				Tenants:  make([]administration.MassAssignmentTenant, 0, room.Capacity),
			})
		}
	}
	return slots
}

func slotsToAssignments(slots []roomSlot) []roomAssignment {
	assignments := make([]roomAssignment, len(slots))
	for i, slot := range slots {
		tenants := make([]administration.MassAssignmentTenant, len(slot.Tenants))
		copy(tenants, slot.Tenants)
		assignments[i] = roomAssignment{
			FlatID:   slot.FlatID,
			RoomID:   slot.RoomID,
			Capacity: slot.Capacity,
			Tenants:  tenants,
		}
	}
	return assignments
}

func totalAssignmentScore(slots []roomSlot, preferenceAttrs []string) int {
	total := 0
	for _, slot := range slots {
		total += roomScore(slot.Tenants, preferenceAttrs)
	}
	return total
}

func cloneRoomSlots(slots []roomSlot) []roomSlot {
	cloned := make([]roomSlot, len(slots))
	for i, slot := range slots {
		cloned[i] = roomSlot{
			FlatID:   slot.FlatID,
			RoomID:   slot.RoomID,
			Capacity: slot.Capacity,
			Tenants:  append([]administration.MassAssignmentTenant(nil), slot.Tenants...),
		}
	}
	return cloned
}

func sortTenantsByPreference(tenants []administration.MassAssignmentTenant, preferenceAttrs []string) []administration.MassAssignmentTenant {
	sorted := append([]administration.MassAssignmentTenant(nil), tenants...)
	sort.Slice(sorted, func(i, j int) bool {
		keyI, err := tenantPreferenceKey(sorted[i], preferenceAttrs)
		if err != nil {
			panic(err)
		}
		keyJ, err := tenantPreferenceKey(sorted[j], preferenceAttrs)
		if err != nil {
			panic(err)
		}
		if keyI != keyJ {
			return keyI < keyJ
		}
		return sorted[i].ID < sorted[j].ID
	})
	return sorted
}

func optimizeTenants(
	tenants []administration.MassAssignmentTenant,
	flats []administration.MassAssignmentFlat,
	preferenceAttrs []string,
) ([]roomAssignment, int) {
	slots := flattenRoomSlots(flats)
	sortedTenants := sortTenantsByPreference(tenants, preferenceAttrs)
	if len(sortedTenants) == 0 {
		return slotsToAssignments(slots), 0
	}
	if len(sortedTenants) > greedyFallbackThreshold {
		bestSlots := optimizeTenantsGreedy(sortedTenants, slots, preferenceAttrs)
		return slotsToAssignments(bestSlots), totalAssignmentScore(bestSlots, preferenceAttrs)
	}
	bestScore := math.MinInt
	bestSlots := cloneRoomSlots(slots)
	backtrackAssignments(sortedTenants, slots, preferenceAttrs, 0, &bestScore, &bestSlots)
	return slotsToAssignments(bestSlots), bestScore
}

func backtrackAssignments(
	tenants []administration.MassAssignmentTenant,
	slots []roomSlot,
	preferenceAttrs []string,
	tenantIdx int,
	bestScore *int,
	bestSlots *[]roomSlot,
) {
	if tenantIdx == len(tenants) {
		score := totalAssignmentScore(slots, preferenceAttrs)
		if score > *bestScore {
			*bestScore = score
			*bestSlots = cloneRoomSlots(slots)
		}
		return
	}

	partialScore := totalAssignmentScore(slots, preferenceAttrs)
	if partialScore <= *bestScore {
		return
	}

	for i := range slots {
		if len(slots[i].Tenants) >= slots[i].Capacity {
			continue
		}
		slots[i].Tenants = append(slots[i].Tenants, tenants[tenantIdx])
		backtrackAssignments(tenants, slots, preferenceAttrs, tenantIdx+1, bestScore, bestSlots)
		slots[i].Tenants = slots[i].Tenants[:len(slots[i].Tenants)-1]
	}
}

func optimizeTenantsGreedy(
	tenants []administration.MassAssignmentTenant,
	slots []roomSlot,
	preferenceAttrs []string,
) []roomSlot {
	remaining := append([]administration.MassAssignmentTenant(nil), tenants...)
	for len(remaining) > 0 {
		bestTenantIdx := -1
		bestSlotIdx := -1
		bestGain := 0
		for tenantIdx, tenant := range remaining {
			for slotIdx := range slots {
				if len(slots[slotIdx].Tenants) >= slots[slotIdx].Capacity {
					continue
				}
				before := roomScore(slots[slotIdx].Tenants, preferenceAttrs)
				slots[slotIdx].Tenants = append(slots[slotIdx].Tenants, tenant)
				after := roomScore(slots[slotIdx].Tenants, preferenceAttrs)
				slots[slotIdx].Tenants = slots[slotIdx].Tenants[:len(slots[slotIdx].Tenants)-1]
				gain := after - before
				if bestTenantIdx == -1 || gain > bestGain {
					bestTenantIdx = tenantIdx
					bestSlotIdx = slotIdx
					bestGain = gain
				}
			}
		}
		slots[bestSlotIdx].Tenants = append(slots[bestSlotIdx].Tenants, remaining[bestTenantIdx])
		remaining = append(remaining[:bestTenantIdx], remaining[bestTenantIdx+1:]...)
	}

	improved := true
	for improved {
		improved = false
		for i := range slots {
			for j := i + 1; j < len(slots); j++ {
				for ti := range slots[i].Tenants {
					for tj := range slots[j].Tenants {
						before := totalAssignmentScore(slots, preferenceAttrs)
						slots[i].Tenants[ti], slots[j].Tenants[tj] = slots[j].Tenants[tj], slots[i].Tenants[ti]
						after := totalAssignmentScore(slots, preferenceAttrs)
						if after > before {
							improved = true
						} else {
							slots[i].Tenants[ti], slots[j].Tenants[tj] = slots[j].Tenants[tj], slots[i].Tenants[ti]
						}
					}
				}
			}
		}
	}
	return slots
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
