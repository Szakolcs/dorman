package administration

import (
	"fmt"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

type MassAssignmentRoomResult struct {
	FlatID   string
	RoomID   string
	Capacity int
	Tenants  []MassAssignmentTenant
}

type massAssignmentRoomSlot struct {
	FlatID   string
	RoomID   string
	Capacity int
	Tenants  []MassAssignmentTenant
}

type massAssignmentCategory struct {
	key  string
	opts [][]int
}

const greedyFallbackThreshold = 12

func parseAttributeList(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func groupTenantsByAttributes(users []MassAssignmentTenant, attributes []string) (map[string][]MassAssignmentTenant, error) {
	if len(users) == 0 || len(attributes) == 0 {
		return make(map[string][]MassAssignmentTenant), nil
	}

	grouped := make(map[string][]MassAssignmentTenant)
	for _, user := range users {
		key, err := tenantPreferenceKey(user, attributes)
		if err != nil {
			return nil, err
		}
		grouped[key] = append(grouped[key], user)
	}
	return grouped, nil
}

func tenantPreferenceKey(tenant MassAssignmentTenant, attributes []string) (string, error) {
	rv := reflect.ValueOf(tenant)
	keyParts := make([]string, 0, len(attributes))
	for _, attr := range attributes {
		fieldName := strings.ToUpper(string(attr[0])) + attr[1:]
		field := rv.FieldByName(fieldName)
		if !field.IsValid() {
			return "", fmt.Errorf("invalid attribute %q: not found in tenant struct", attr)
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

func roomScore(tenants []MassAssignmentTenant, preferenceAttrs []string) int {
	if len(tenants) < 2 {
		return 0
	}
	keys := make([]string, len(tenants))
	for i, tenant := range tenants {
		key, err := tenantPreferenceKey(tenant, preferenceAttrs)
		if err != nil {
			return 0
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

// RoomCompatibilityScore returns the pairwise preference score for tenants in one room.
func RoomCompatibilityScore(tenants []MassAssignmentTenant, preferenceAttrs []string) int {
	return roomScore(tenants, preferenceAttrs)
}

func flattenMassAssignmentRoomSlots(flats []MassAssignmentFlat) []massAssignmentRoomSlot {
	slots := make([]massAssignmentRoomSlot, 0)
	for _, flat := range flats {
		for _, room := range flat.Rooms {
			slots = append(slots, massAssignmentRoomSlot{
				FlatID:   flat.ID,
				RoomID:   room.ID,
				Capacity: room.Capacity,
				Tenants:  make([]MassAssignmentTenant, 0, room.Capacity),
			})
		}
	}
	return slots
}

func slotsToMassAssignmentResults(slots []massAssignmentRoomSlot) []MassAssignmentRoomResult {
	results := make([]MassAssignmentRoomResult, len(slots))
	for i, slot := range slots {
		tenants := make([]MassAssignmentTenant, len(slot.Tenants))
		copy(tenants, slot.Tenants)
		results[i] = MassAssignmentRoomResult{
			FlatID:   slot.FlatID,
			RoomID:   slot.RoomID,
			Capacity: slot.Capacity,
			Tenants:  tenants,
		}
	}
	return results
}

func totalAssignmentScore(slots []massAssignmentRoomSlot, preferenceAttrs []string) int {
	total := 0
	for _, slot := range slots {
		total += roomScore(slot.Tenants, preferenceAttrs)
	}
	return total
}

func cloneMassAssignmentRoomSlots(slots []massAssignmentRoomSlot) []massAssignmentRoomSlot {
	cloned := make([]massAssignmentRoomSlot, len(slots))
	for i, slot := range slots {
		cloned[i] = massAssignmentRoomSlot{
			FlatID:   slot.FlatID,
			RoomID:   slot.RoomID,
			Capacity: slot.Capacity,
			Tenants:  append([]MassAssignmentTenant(nil), slot.Tenants...),
		}
	}
	return cloned
}

func sortTenantsByPreference(tenants []MassAssignmentTenant, preferenceAttrs []string) []MassAssignmentTenant {
	sorted := append([]MassAssignmentTenant(nil), tenants...)
	sort.Slice(sorted, func(i, j int) bool {
		keyI, err := tenantPreferenceKey(sorted[i], preferenceAttrs)
		if err != nil {
			return sorted[i].ID < sorted[j].ID
		}
		keyJ, err := tenantPreferenceKey(sorted[j], preferenceAttrs)
		if err != nil {
			return sorted[i].ID < sorted[j].ID
		}
		if keyI != keyJ {
			return keyI < keyJ
		}
		return sorted[i].ID < sorted[j].ID
	})
	return sorted
}

func optimizeTenantRoomAssignments(
	tenants []MassAssignmentTenant,
	flats []MassAssignmentFlat,
	preferenceAttrs []string,
) ([]MassAssignmentRoomResult, int) {
	slots := flattenMassAssignmentRoomSlots(flats)
	sortedTenants := sortTenantsByPreference(tenants, preferenceAttrs)
	if len(sortedTenants) == 0 {
		return slotsToMassAssignmentResults(slots), 0
	}
	if len(sortedTenants) > greedyFallbackThreshold {
		bestSlots := optimizeTenantRoomAssignmentsGreedy(sortedTenants, slots, preferenceAttrs)
		return slotsToMassAssignmentResults(bestSlots), totalAssignmentScore(bestSlots, preferenceAttrs)
	}
	bestScore := math.MinInt
	bestSlots := cloneMassAssignmentRoomSlots(slots)
	backtrackTenantRoomAssignments(sortedTenants, slots, preferenceAttrs, 0, &bestScore, &bestSlots)
	return slotsToMassAssignmentResults(bestSlots), bestScore
}

func backtrackTenantRoomAssignments(
	tenants []MassAssignmentTenant,
	slots []massAssignmentRoomSlot,
	preferenceAttrs []string,
	tenantIdx int,
	bestScore *int,
	bestSlots *[]massAssignmentRoomSlot,
) {
	if tenantIdx == len(tenants) {
		score := totalAssignmentScore(slots, preferenceAttrs)
		if score > *bestScore {
			*bestScore = score
			*bestSlots = cloneMassAssignmentRoomSlots(slots)
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
		backtrackTenantRoomAssignments(tenants, slots, preferenceAttrs, tenantIdx+1, bestScore, bestSlots)
		slots[i].Tenants = slots[i].Tenants[:len(slots[i].Tenants)-1]
	}
}

func optimizeTenantRoomAssignmentsGreedy(
	tenants []MassAssignmentTenant,
	slots []massAssignmentRoomSlot,
	preferenceAttrs []string,
) []massAssignmentRoomSlot {
	remaining := append([]MassAssignmentTenant(nil), tenants...)
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
	coinChangeDepthFirst(0, amount, path, &result, coins)
	return result
}

func coinChangeDepthFirst(startIdx int, remaining int, path []int, result *[][]int, coins []int) {
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
		coinChangeDepthFirst(i+1, remaining-coins[i], path, result, coins)
		path = path[:len(path)-1]
	}
}

func findValidFlatArrangement(coins []int, subsets map[string][][]int) (map[string][]int, bool) {
	if len(subsets) == 0 {
		return nil, false
	}

	origFreq := make(map[int]int)
	for _, coin := range coins {
		origFreq[coin]++
	}

	cats := make([]massAssignmentCategory, 0, len(subsets))
	for key, options := range subsets {
		cats = append(cats, massAssignmentCategory{key, options})
	}
	sort.Slice(cats, func(i, j int) bool {
		return len(cats[i].opts) < len(cats[j].opts)
	})

	result := make(map[string][]int)
	currFreq := make(map[int]int, len(origFreq))
	for key, value := range origFreq {
		currFreq[key] = value
	}
	if flatArrangementSolver(0, currFreq, cats, &result) {
		return result, true
	}
	return nil, false
}

func flatArrangementSolver(idx int, freq map[int]int, cats []massAssignmentCategory, result *map[string][]int) bool {
	if idx == len(cats) {
		return true
	}
	key := cats[idx].key
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
			(*result)[key] = candidate
			if flatArrangementSolver(idx+1, freq, cats, result) {
				return true
			}
			delete(*result, key)
		}

		for _, coin := range used {
			freq[coin]++
		}
	}
	return false
}

func sumRoomCapacities(flats []MassAssignmentFlat) int {
	total := 0
	for _, flat := range flats {
		for _, room := range flat.Rooms {
			total += room.Capacity
		}
	}
	return total
}

// RunMassAssignment assigns tenants to rooms using strict groups, flat allocation,
// and preference-based room optimization.
func RunMassAssignment(
	req TenantMassAssignmentRequest,
	tenants []MassAssignmentTenant,
	flats []MassAssignmentFlat,
) ([]MassAssignmentRoomResult, error) {
	strictAttrs := parseAttributeList(req.StrictGroups)
	preferenceAttrs := parseAttributeList(req.Preferences)
	if len(strictAttrs) == 0 {
		return nil, fmt.Errorf("%w: strict groups are required", ErrValidation)
	}
	if len(preferenceAttrs) == 0 {
		return nil, fmt.Errorf("%w: preferences are required", ErrValidation)
	}
	if len(tenants) == 0 {
		return nil, fmt.Errorf("%w: no tenants to assign", ErrValidation)
	}
	if len(flats) == 0 {
		return nil, fmt.Errorf("%w: no flats available", ErrValidation)
	}

	for i := range flats {
		if flats[i].Capacity == 0 {
			for _, room := range flats[i].Rooms {
				flats[i].Capacity += room.Capacity
			}
		}
	}

	groupedTenants, err := groupTenantsByAttributes(tenants, strictAttrs)
	if err != nil {
		return nil, err
	}

	flatCapacities := make([]int, 0, len(flats))
	availableFlats := append([]MassAssignmentFlat(nil), flats...)
	for _, flat := range availableFlats {
		flatCapacities = append(flatCapacities, flat.Capacity)
	}

	coinChanges := make(map[string][][]int)
	for key, group := range groupedTenants {
		coinChanges[key] = calculateCoinChange(flatCapacities, len(group))
	}

	validArrangement, ok := findValidFlatArrangement(flatCapacities, coinChanges)
	if !ok {
		return nil, fmt.Errorf("%w: no valid flat arrangement found for strict groups", ErrValidation)
	}

	strictGroupKeys := make([]string, 0, len(groupedTenants))
	for key := range groupedTenants {
		strictGroupKeys = append(strictGroupKeys, key)
	}
	sort.Strings(strictGroupKeys)

	validArrangementToFlats := make(map[string][]MassAssignmentFlat)
	for groupKey, sizes := range validArrangement {
		for _, size := range sizes {
			matched := false
			for i, flat := range availableFlats {
				if flat.Capacity == size {
					validArrangementToFlats[groupKey] = append(validArrangementToFlats[groupKey], flat)
					availableFlats = append(availableFlats[:i], availableFlats[i+1:]...)
					matched = true
					break
				}
			}
			if !matched {
				return nil, fmt.Errorf("%w: could not map flat capacity %d for strict group %q", ErrValidation, size, groupKey)
			}
		}
	}

	allResults := make([]MassAssignmentRoomResult, 0)
	for _, groupKey := range strictGroupKeys {
		groupTenants := groupedTenants[groupKey]
		groupFlats := validArrangementToFlats[groupKey]
		roomCapTotal := sumRoomCapacities(groupFlats)
		if roomCapTotal != len(groupTenants) {
			return nil, fmt.Errorf(
				"%w: strict group %q room capacity %d does not match tenant count %d",
				ErrValidation,
				groupKey,
				roomCapTotal,
				len(groupTenants),
			)
		}

		assignments, _ := optimizeTenantRoomAssignments(groupTenants, groupFlats, preferenceAttrs)
		allResults = append(allResults, assignments...)
	}

	return allResults, nil
}
