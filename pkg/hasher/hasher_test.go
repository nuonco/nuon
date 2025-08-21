package hasher

import (
	"testing"
)

type TestStruct struct {
	ID       int    `nuonhash:"id"`
	Name     string `nuonhash:"name"`
	Password string `nuonhash:"-"`
	Age      int    `nuonhash:"age"`
}

type MapTestStruct struct {
	Name      string            `nuonhash:"name"`
	ValuesMap map[string]string `nuonhash:"values"`
	Secret    string            `nuonhash:"-"`
}

func TestHashStructBasic(t *testing.T) {
	t.Run("ignored_fields_dont_affect_hash", func(t *testing.T) {
		s1 := TestStruct{ID: 1, Name: "test", Password: "secret1", Age: 25}
		s2 := TestStruct{ID: 1, Name: "test", Password: "secret2", Age: 25} // Different ignored field

		hash1, err1 := HashStruct(s1)
		if err1 != nil {
			t.Fatalf("HashStruct failed for s1: %v", err1)
		}

		hash2, err2 := HashStruct(s2)
		if err2 != nil {
			t.Fatalf("HashStruct failed for s2: %v", err2)
		}

		if hash1 != hash2 {
			t.Errorf("Expected hashes to match when only ignored fields differ")
			t.Errorf("Hash1: %s", hash1)
			t.Errorf("Hash2: %s", hash2)
		}

		t.Logf("✅ Hashes match correctly (ignored field different): %s", hash1[:16]+"...")
	})

	t.Run("included_fields_affect_hash", func(t *testing.T) {
		s1 := TestStruct{ID: 1, Name: "test", Password: "secret", Age: 25}
		s2 := TestStruct{ID: 1, Name: "different", Password: "secret", Age: 25} // Different included field

		hash1, err1 := HashStruct(s1)
		if err1 != nil {
			t.Fatalf("HashStruct failed for s1: %v", err1)
		}

		hash2, err2 := HashStruct(s2)
		if err2 != nil {
			t.Fatalf("HashStruct failed for s2: %v", err2)
		}

		if hash1 == hash2 {
			t.Errorf("Expected hashes to differ when included fields differ")
			t.Errorf("Hash1: %s", hash1)
			t.Errorf("Hash2: %s", hash2)
		}

		t.Logf("✅ Hashes differ correctly when included fields change")
	})
}

func TestHashStructMapConsistency(t *testing.T) {
	t.Run("map_same_data_different_insertion_order", func(t *testing.T) {
		m1 := MapTestStruct{
			Name: "test",
			ValuesMap: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
			Secret: "secret1",
		}

		m2 := MapTestStruct{
			Name: "test",
			ValuesMap: map[string]string{
				"key3": "value3", // Different insertion order
				"key1": "value1",
				"key2": "value2",
			},
			Secret: "secret2", // Different ignored field
		}

		// Test multiple times to check consistency
		for i := 0; i < 5; i++ {
			hashM1, err := HashStruct(m1)
			if err != nil {
				t.Fatalf("HashStruct failed for m1 on run %d: %v", i+1, err)
			}

			hashM2, err := HashStruct(m2)
			if err != nil {
				t.Fatalf("HashStruct failed for m2 on run %d: %v", i+1, err)
			}

			if hashM1 != hashM2 {
				t.Errorf("Run %d: Maps with same data produced different hashes", i+1)
				t.Errorf("Map1 hash: %s", hashM1)
				t.Errorf("Map2 hash: %s", hashM2)
				t.Fatal("❌ INCONSISTENT: Maps with same data should always produce same hash")
			}

			t.Logf("Run %d: ✅ Consistent hashes: %s", i+1, hashM1[:16]+"...")
		}
	})

	t.Run("map_different_data_different_hash", func(t *testing.T) {
		m1 := MapTestStruct{
			Name: "test",
			ValuesMap: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		}

		m2 := MapTestStruct{
			Name: "test",
			ValuesMap: map[string]string{
				"key1": "value1",
				"key2": "different_value", // Different value
			},
		}

		hash1, err1 := HashStruct(m1)
		if err1 != nil {
			t.Fatalf("HashStruct failed for m1: %v", err1)
		}

		hash2, err2 := HashStruct(m2)
		if err2 != nil {
			t.Fatalf("HashStruct failed for m2: %v", err2)
		}

		if hash1 == hash2 {
			t.Errorf("Expected different hashes for maps with different data")
			t.Errorf("Hash1: %s", hash1)
			t.Errorf("Hash2: %s", hash2)
		}

		t.Logf("✅ Different map data produces different hashes")
	})
}

