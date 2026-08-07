package common

import (
	"testing"
)

func TestNewStringSet(t *testing.T) {
	testCases := []struct {
		name          string
		inputs        []string
		expectedElems []string
	}{
		{
			name:          "single string, single set",
			inputs:        []string{"hello1"},
			expectedElems: []string{"hello1"},
		},
		{
			name:          "empty set",
			inputs:        []string{},
			expectedElems: []string{},
		},
		{
			name:          "multiple strings",
			inputs:        []string{"a", "b", "c"},
			expectedElems: []string{"a", "b", "c"},
		},
		{
			name:          "duplicates collapse",
			inputs:        []string{"a", "a", "b"},
			expectedElems: []string{"a", "b"},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			set := NewSet(tt.inputs...)

			if set.Len() != len(tt.expectedElems) {
				t.Fatalf("Expected len %d but got %d", len(tt.expectedElems), set.Len())
			}

			for _, e := range tt.expectedElems {
				if !set.Contains(e) {
					t.Fatalf("Expected set to contain %q but it did not: %s", e, set.String())
				}
			}
		})
	}
}

func TestAddString(t *testing.T) {
	testCases := []struct {
		name     string
		setup    *Set[string]
		add      string
		expected *Set[string]
	}{
		{
			name:     "add to nil map via zero-value Set",
			setup:    &Set[string]{},
			add:      "a",
			expected: NewSet("a"),
		},
		{
			name:     "add new element to existing set",
			setup:    NewSet("a"),
			add:      "b",
			expected: NewSet("a", "b"),
		},
		{
			name:     "add duplicate is a no-op",
			setup:    NewSet("a", "b"),
			add:      "a",
			expected: NewSet("a", "b"),
		},
		{
			name:     "add second element to single-element set",
			setup:    NewSet("a"),
			add:      "b",
			expected: NewSet("a", "b"),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.setup.Add(tt.add)

			if !tt.setup.Equal(tt.expected) {
				t.Fatalf("Expected %s but got %s", tt.expected.String(), tt.setup.String())
			}
		})
	}
}

func TestRemoveString(t *testing.T) {
	testCases := []struct {
		name     string
		setup    *Set[string]
		remove   string
		expected *Set[string]
	}{
		{
			name:     "remove from nil map via zero-value Set is a no-op",
			setup:    &Set[string]{},
			remove:   "a",
			expected: &Set[string]{},
		},
		{
			name:     "remove only element leaves empty set",
			setup:    NewSet("a"),
			remove:   "a",
			expected: NewSet[string](),
		},
		{
			name:     "remove one of several elements",
			setup:    NewSet("a", "b", "c"),
			remove:   "b",
			expected: NewSet("a", "c"),
		},
		{
			name:     "remove element not in set is a no-op",
			setup:    NewSet("a", "b"),
			remove:   "z",
			expected: NewSet("a", "b"),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.setup.Remove(tt.remove)

			if !tt.setup.Equal(tt.expected) {
				t.Fatalf("Expected %s but got %s", tt.expected.String(), tt.setup.String())
			}
		})
	}
}

func TestAll(t *testing.T) {
	testCases := []struct {
		name     string
		setup    *Set[string]
		expected *Set[string]
	}{
		{
			name:     "nil map via zero-value Set yields nothing",
			setup:    &Set[string]{},
			expected: NewSet[string](),
		},
		{
			name:     "empty set yields nothing",
			setup:    NewSet[string](),
			expected: NewSet[string](),
		},
		{
			name:     "single element",
			setup:    NewSet("a"),
			expected: NewSet("a"),
		},
		{
			name:     "multiple elements",
			setup:    NewSet("a", "b", "c"),
			expected: NewSet("a", "b", "c"),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			set := NewSet[string]()
			for elem := range tt.setup.All() {
				set.Add(elem)
			}

			if !set.Equal(tt.expected) {
				t.Fatalf("Expected %s but got %s", tt.expected.String(), set.String())
			}
		})
	}
}

func TestEqual(t *testing.T) {
	testCases := []struct {
		name            string
		a               *Set[string]
		b               *Set[string]
		expectToBeEqual bool
	}{
		{name: "both nil", a: nil, b: nil, expectToBeEqual: true},
		{name: "a nil b empty", a: nil, b: NewSet[string](), expectToBeEqual: false},
		{name: "a empty b nil", a: NewSet[string](), b: nil, expectToBeEqual: false},
		{name: "both empty", a: NewSet[string](), b: NewSet[string](), expectToBeEqual: true},
		{name: "same single element", a: NewSet("a"), b: NewSet("a"), expectToBeEqual: true},
		{
			name:            "same multiple elements",
			a:               NewSet("a", "b", "c"),
			b:               NewSet("a", "b", "c"),
			expectToBeEqual: true,
		},
		{
			name:            "different order same elements",
			a:               NewSet("a", "b", "c"),
			b:               NewSet("c", "b", "a"),
			expectToBeEqual: true,
		},
		{name: "different single element", a: NewSet("a"), b: NewSet("b"), expectToBeEqual: false},
		{name: "a subset of b", a: NewSet("a"), b: NewSet("a", "b"), expectToBeEqual: false},
		{name: "b subset of a", a: NewSet("a", "b"), b: NewSet("a"), expectToBeEqual: false},
		{name: "disjoint", a: NewSet("a"), b: NewSet("b"), expectToBeEqual: false},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if isEqual := tt.a.Equal(tt.b); isEqual != tt.expectToBeEqual {
				t.Fatalf("Equal(%s, %s) = %v, want %v", tt.a, tt.b, isEqual, tt.expectToBeEqual)
			}
		})
	}
}

func TestContains(t *testing.T) {
	setup := NewSet("a", "b", "c")

	testCases := []struct {
		name          string
		set           *Set[string]
		item          string
		shouldContain bool
	}{
		{name: "present first", set: setup, item: "a", shouldContain: true},
		{name: "present middle", set: setup, item: "b", shouldContain: true},
		{name: "present last", set: setup, item: "c", shouldContain: true},
		{name: "absent", set: setup, item: "z", shouldContain: false},
		{name: "empty string absent from non-empty", set: setup, item: "", shouldContain: false},
		{name: "absent from empty set", set: NewSet[string](), item: "a", shouldContain: false},
		{
			name:          "absent from nil-map zero-value set",
			set:           &Set[string]{},
			item:          "a",
			shouldContain: false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if doesContain := tt.set.Contains(tt.item); doesContain != tt.shouldContain {
				t.Fatalf("Contains(%q) = %v, want %v", tt.item, doesContain, tt.shouldContain)
			}
		})
	}
}

func TestLen(t *testing.T) {
	testCases := []struct {
		name        string
		Set         *Set[string]
		expectedLen int
	}{
		{name: "nil map via zero-value Set", Set: &Set[string]{}, expectedLen: 0},
		{name: "empty set", Set: NewSet[string](), expectedLen: 0},
		{name: "single element", Set: NewSet("a"), expectedLen: 1},
		{name: "multiple elements", Set: NewSet("a", "b", "c"), expectedLen: 3},
		{name: "duplicates collapse", Set: NewSet("a", "a", "b", "b"), expectedLen: 2},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if actualLen := tt.Set.Len(); actualLen != tt.expectedLen {
				t.Fatalf("Len() = %d, want %d", actualLen, tt.expectedLen)
			}
		})
	}
}
