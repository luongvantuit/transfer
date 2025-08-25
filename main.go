package main

import (
	"crypto/aes"
	crand "crypto/rand"
	"fmt"
	mrand "math/rand"
	"os"
	"strings"
	"time"

	"github.com/luongvantuit/transfer/cipher"
)

// Old encryption functions removed - now using SubstitutionCipher from cipher package

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Generate valid AES key for FPE
func mustAESKey() []byte {
	key := make([]byte, 32) // AES-256
	if _, err := crand.Read(key); err != nil {
		panic(err)
	}
	// ensure it's a valid AES key
	if _, err := aes.NewCipher(key); err != nil {
		panic(err)
	}
	return key
}

func generateUniqueNumbers(count int) []string {
	used := make(map[string]bool)
	numbers := make([]string, 0, count)

	for len(numbers) < count {
		// Generate numbers from 1-999999
		num := mrand.Intn(999999) + 1
		numStr := fmt.Sprintf("%d", num)

		if !used[numStr] {
			used[numStr] = true
			numbers = append(numbers, numStr)
		}
	}

	return numbers
}

func generateUniqueStrings(count int) []string {
	used := make(map[string]bool)
	strings := make([]string, 0, count)
	// const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+-=[]{}|;:,.<>?"
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	for len(strings) < count {
		// Generate string with length from 1-20 characters
		length := mrand.Intn(20) + 1
		result := make([]byte, length)

		for i := range result {
			result[i] = charset[mrand.Intn(len(charset))]
		}

		str := string(result)
		if !used[str] {
			used[str] = true
			strings = append(strings, str)
		}
	}

	return strings
}

func createASCIIChart(title string, data map[string]float64, width int) string {
	result := fmt.Sprintf("\n%s\n", title)
	result += strings.Repeat("=", len(title)) + "\n\n"

	maxValue := 0.0
	for _, v := range data {
		if v > maxValue {
			maxValue = v
		}
	}

	for label, value := range data {
		barLength := int((value / maxValue) * float64(width))
		bar := strings.Repeat("█", barLength)
		result += fmt.Sprintf("%-20s |%s| %.2f\n", label, bar, value)
	}

	return result
}

func printPerformanceSummary(numbersTime, stringsTime, fpeTime time.Duration, testCount int) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("                    PERFORMANCE SUMMARY")
	fmt.Println(strings.Repeat("=", 60))

	// Numbers performance
	numbersPerSec := float64(testCount) / numbersTime.Seconds()
	numbersPerMs := float64(testCount) / float64(numbersTime.Milliseconds())

	// Strings performance
	stringsPerSec := float64(testCount) / stringsTime.Seconds()
	stringsPerMs := float64(testCount) / float64(stringsTime.Milliseconds())

	fmt.Printf("Numbers Processing:\n")
	fmt.Printf("  Total time: %v\n", numbersTime)
	fmt.Printf("  Rate: %.0f items/sec (%.0f items/ms)\n", numbersPerSec, float64(numbersPerMs))
	fmt.Printf("  Average: %.3f μs per item\n", float64(numbersTime.Microseconds())/1000000)

	fmt.Printf("\nStrings Processing:\n")
	fmt.Printf("  Total time: %v\n", stringsTime)
	fmt.Printf("  Rate: %.0f items/sec (%.0f items/ms)\n", stringsPerSec, float64(stringsPerMs))
	fmt.Printf("  Average: %.3f μs per item\n", float64(stringsTime.Microseconds())/1000000)

	// FPE performance
	fpePerSec := float64(testCount) / fpeTime.Seconds()
	fpePerMs := float64(testCount) / float64(fpeTime.Milliseconds())

	fmt.Printf("\nFPE Cipher Processing:\n")
	fmt.Printf("  Total time: %v\n", fpeTime)
	fmt.Printf("  Rate: %.0f items/sec (%.0f items/ms)\n", fpePerSec, float64(fpePerMs))
	fmt.Printf("  Average: %.3f μs per item\n", float64(fpeTime.Microseconds())/1000000)

	// Performance comparison
	fastest := "Numbers"
	fastestTime := numbersTime
	if stringsTime < fastestTime {
		fastest = "Strings"
		fastestTime = stringsTime
	}
	if fpeTime < fastestTime {
		fastest = "FPE Cipher"
		fastestTime = fpeTime
	}

	fmt.Printf("\n%s is the fastest\n", fastest)
	if fastestTime != numbersTime {
		fmt.Printf("Numbers are %.1fx slower than %s\n", float64(numbersTime)/float64(fastestTime), fastest)
	}
	if fastestTime != stringsTime {
		fmt.Printf("Strings are %.1fx slower than %s\n", float64(stringsTime)/float64(fastestTime), fastest)
	}
	if fastestTime != fpeTime {
		fmt.Printf("FPE Cipher is %.1fx slower than %s\n", float64(fpeTime)/float64(fastestTime), fastest)
	}
}

func main() {
	key := "IhlVHM9D4N1B2vVDd4QAgdiJ3zh60L1q"

	// Configuration - Easy to change
	const testCount = 100000 // Change this value to adjust test size
	const sampleCount = 5000 // Change this value to adjust number of samples displayed

	// Initialize SubstitutionCipher
	subCipher := cipher.NewSubstitutionCipher(key)

	// Initialize FPE Cipher (FF1)
	fpeKey := mustAESKey() // Generate valid AES-256 key
	fpeCipher, err := cipher.NewFPECipher(fpeKey)
	if err != nil {
		fmt.Printf("Error creating FPE cipher: %v\n", err)
		return
	}

	// Basic test
	plain := "69619"
	encrypted := subCipher.Encrypt(plain)
	decrypted := subCipher.Decrypt(encrypted)

	fmt.Println("=== BASIC TEST ===")
	fmt.Println("key:     ", key)
	fmt.Println("plain:   ", plain)
	fmt.Println("encrypted:", encrypted)
	fmt.Println("decoded: ", decrypted)

	// Test EncryptNumber/DecryptNumber (special handling for numbers)
	numberPlain := "12345"
	numberEncrypted := subCipher.EncryptNumber(numberPlain)
	numberDecrypted := subCipher.DecryptNumber(numberEncrypted)

	fmt.Println("\n=== NUMBER-SPECIFIC TEST ===")
	fmt.Println("number plain:   ", numberPlain)
	fmt.Println("number encrypted:", numberEncrypted)
	fmt.Println("number decoded: ", numberDecrypted)

	// Test FPE Cipher
	fpePlain := "12345" // Simple number for FF1 testing
	fpeEncrypted, err := fpeCipher.EncryptPreserving(fpePlain)
	if err != nil {
		fmt.Printf("Error encrypting with FPE: %v\n", err)
		return
	}
	fpeDecrypted, err := fpeCipher.DecryptPreserving(fpeEncrypted)
	if err != nil {
		fmt.Printf("Error decrypting with FPE: %v\n", err)
		return
	}

	fmt.Println("\n=== FPE CIPHER TEST ===")
	fmt.Println("FPE plain:     ", fpePlain)
	fmt.Println("FPE encrypted: ", fpeEncrypted)
	fmt.Println("FPE decoded:   ", fpeDecrypted)
	fmt.Println()

	// Benchmark with unique random numbers
	fmt.Printf("Generating %d unique random numbers...\n", testCount)
	start := time.Now()

	numbers := generateUniqueNumbers(testCount)
	fmt.Printf("Generated %d unique numbers in %v\n", len(numbers), time.Since(start))

	// Encrypt and decrypt with progress tracking
	fmt.Println("Starting encryption/decryption of numbers...")
	encryptedNumbers := make([]string, testCount)
	decryptedNumbers := make([]string, testCount)

	batchSize := testCount / 10 // Progress every 10%
	correctNumbers := 0

	for i := 0; i < testCount; i++ {
		encryptedNumbers[i] = subCipher.EncryptNumber(numbers[i])
		decryptedNumbers[i] = subCipher.DecryptNumber(encryptedNumbers[i])

		// Verify and track progress
		if numbers[i] == decryptedNumbers[i] {
			correctNumbers++
		}

		// Progress report every batch
		if (i+1)%batchSize == 0 {
			progress := float64(i+1) / float64(testCount) * 100
			elapsed := time.Since(start)
			rate := float64(i+1) / elapsed.Seconds()
			fmt.Printf("Progress: %.1f%% (%d/%d) - Rate: %.0f items/sec - Elapsed: %v\n",
				progress, i+1, testCount, rate, elapsed)
		}
	}

	numbersTime := time.Since(start)

	// Test with unique random strings
	fmt.Printf("Generating %d unique random strings...\n", testCount)
	start = time.Now()

	strings := generateUniqueStrings(testCount)
	fmt.Printf("Generated %d unique strings in %v\n", len(strings), time.Since(start))

	// Encrypt and decrypt with progress tracking
	fmt.Println("Starting encryption/decryption of strings...")
	encryptedStrings := make([]string, testCount)
	decryptedStrings := make([]string, testCount)

	correctStrings := 0

	for i := 0; i < testCount; i++ {
		encryptedStrings[i] = subCipher.Encrypt(strings[i])
		decryptedStrings[i] = subCipher.Decrypt(encryptedStrings[i])

		// Verify and track progress
		if strings[i] == decryptedStrings[i] {
			correctStrings++
		}

		// Progress report every batch
		if (i+1)%batchSize == 0 {
			progress := float64(i+1) / float64(testCount) * 100
			elapsed := time.Since(start)
			rate := float64(i+1) / elapsed.Seconds()
			fmt.Printf("Progress: %.1f%% (%d/%d) - Rate: %.0f items/sec - Elapsed: %v\n",
				progress, i+1, testCount, rate, elapsed)
		}
	}

	stringsTime := time.Since(start)

	// Check for duplicates in outputs
	fmt.Println("\n=== DUPLICATE CHECK ===")

	// Check encrypted numbers for duplicates
	encryptedNumbersSet := make(map[string]bool)
	duplicateNumbers := 0
	for _, enc := range encryptedNumbers {
		if encryptedNumbersSet[enc] {
			duplicateNumbers++
		} else {
			encryptedNumbersSet[enc] = true
		}
	}
	fmt.Printf("Encrypted numbers: %d unique, %d duplicates (%.2f%%)\n",
		len(encryptedNumbersSet), duplicateNumbers, float64(duplicateNumbers)/float64(testCount)*100)

	// Check encrypted strings for duplicates
	encryptedStringsSet := make(map[string]bool)
	duplicateStrings := 0
	for _, enc := range encryptedStrings {
		if encryptedStringsSet[enc] {
			duplicateStrings++
		} else {
			encryptedStringsSet[enc] = true
		}
	}
	fmt.Printf("Encrypted strings: %d unique, %d duplicates (%.2f%%)\n",
		len(encryptedStringsSet), duplicateStrings, float64(duplicateStrings)/float64(testCount)*100)

	// Check if any encrypted output matches input
	inputOutputCollision := 0
	for i := 0; i < testCount; i++ {
		if numbers[i] == encryptedNumbers[i] {
			inputOutputCollision++
		}
	}
	fmt.Printf("Input-Output collisions: %d (%.2f%%)\n",
		inputOutputCollision, float64(inputOutputCollision)/float64(testCount)*100)

	// Benchmark FPE Cipher
	fmt.Println("\n=== FPE CIPHER BENCHMARK ===")
	fmt.Printf("Testing FPE Cipher with %d mixed strings...\n", testCount)

	start = time.Now()
	fpeEncryptedStrings := make([]string, testCount)
	fpeDecryptedStrings := make([]string, testCount)
	correctFPE := 0

	for i := 0; i < testCount; i++ {
		// Create simple numbers for FPE testing (FF1 works best with digits)
		fpeTestString := fmt.Sprintf("%d", 100000+i) // Generate numbers 100000-199999

		fpeEncryptedStrings[i], err = fpeCipher.EncryptPreserving(fpeTestString)
		if err != nil {
			fmt.Printf("Error encrypting with FPE: %v\n", err)
			return
		}

		fpeDecryptedStrings[i], err = fpeCipher.DecryptPreserving(fpeEncryptedStrings[i])
		if err != nil {
			fmt.Printf("Error decrypting with FPE: %v\n", err)
			return
		}

		if fpeTestString == fpeDecryptedStrings[i] {
			correctFPE++
		}

		// Progress report every batch
		if (i+1)%batchSize == 0 {
			progress := float64(i+1) / float64(testCount) * 100
			elapsed := time.Since(start)
			rate := float64(i+1) / elapsed.Seconds()
			fmt.Printf("FPE Progress: %.1f%% (%d/%d) - Rate: %.0f items/sec - Elapsed: %v\n",
				progress, i+1, testCount, rate, elapsed)
		}
	}

	fpeTime := time.Since(start)
	fmt.Printf("FPE Cipher completed in %v\n", fpeTime)
	fmt.Printf("FPE Accuracy: %d/%d (%.2f%%)\n", correctFPE, testCount, float64(correctFPE)/float64(testCount)*100)

	// Write test data to test.txt
	testFile, err := os.Create("test.txt")
	if err != nil {
		fmt.Printf("Error creating test.txt: %v\n", err)
		return
	}
	defer testFile.Close()

	// Write numbers to test.txt (pure data only)
	for i := 0; i < testCount; i++ {
		fmt.Fprintf(testFile, "%s\n", numbers[i])
	}

	// Write strings to test.txt (pure data only)
	for i := 0; i < testCount; i++ {
		fmt.Fprintf(testFile, "%s\n", strings[i])
	}

	// Write results to out.txt
	file, err := os.Create("out.txt")
	if err != nil {
		fmt.Printf("Error creating out.txt: %v\n", err)
		return
	}
	defer file.Close()

	fmt.Fprintf(file, "=== BENCHMARK RESULTS ===\n\n")
	fmt.Fprintf(file, "Key: %s\n\n", key)

	fmt.Fprintf(file, "=== NUMBERS TEST (%d unique numbers) ===\n", testCount)
	fmt.Fprintf(file, "Time taken: %v\n", numbersTime)
	fmt.Fprintf(file, "Correct decrypts: %d/%d (%.2f%%)\n", correctNumbers, testCount, float64(correctNumbers)/float64(testCount)*100)
	fmt.Fprintf(file, "Average time per number: %v\n\n", numbersTime/time.Duration(testCount))

	fmt.Fprintf(file, "=== STRINGS TEST (%d unique strings) ===\n", testCount)
	fmt.Fprintf(file, "Time taken: %v\n", stringsTime)
	fmt.Fprintf(file, "Correct decrypts: %d/%d (%.2f%%)\n", correctStrings, testCount, float64(correctStrings)/float64(testCount)*100)
	fmt.Fprintf(file, "Average time per string: %v\n\n", stringsTime/time.Duration(testCount))

	fmt.Fprintf(file, "=== FPE CIPHER TEST (%d mixed strings) ===\n", testCount)
	fmt.Fprintf(file, "Time taken: %v\n", fpeTime)
	fmt.Fprintf(file, "Correct decrypts: %d/%d (%.2f%%)\n", correctFPE, testCount, float64(correctFPE)/float64(testCount)*100)
	fmt.Fprintf(file, "Average time per string: %v\n\n", fpeTime/time.Duration(testCount))

	fmt.Fprintf(file, "=== SAMPLE RESULTS ===\n")
	fmt.Fprintf(file, "First %d numbers:\n", sampleCount)
	for i := 0; i < sampleCount; i++ {
		fmt.Fprintf(file, "  Input: %s -> Encrypted: %s -> Decrypted: %s (Match: %t)\n",
			numbers[i], encryptedNumbers[i], decryptedNumbers[i], numbers[i] == decryptedNumbers[i])
	}

	fmt.Fprintf(file, "\nFirst %d strings:\n", sampleCount)
	for i := 0; i < sampleCount; i++ {
		fmt.Fprintf(file, "  Input: %s -> Encrypted: %s -> Decrypted: %s (Match: %t)\n",
			strings[i], encryptedStrings[i], decryptedStrings[i], strings[i] == decryptedStrings[i])
	}

	fmt.Fprintf(file, "\n=== PERFORMANCE STATISTICS ===\n")
	fmt.Fprintf(file, "Numbers Processing:\n")
	fmt.Fprintf(file, "  Total time: %v\n", numbersTime)
	fmt.Fprintf(file, "  Rate: %.0f items/sec\n", float64(testCount)/numbersTime.Seconds())
	fmt.Fprintf(file, "  Average: %.3f μs per item\n", float64(numbersTime.Microseconds())/float64(testCount))

	fmt.Fprintf(file, "\nStrings Processing:\n")
	fmt.Fprintf(file, "  Total time: %v\n", stringsTime)
	fmt.Fprintf(file, "  Rate: %.0f items/sec\n", float64(testCount)/stringsTime.Seconds())
	fmt.Fprintf(file, "  Average: %.3f μs per item\n", float64(stringsTime.Microseconds())/float64(testCount))

	fmt.Fprintf(file, "\nFPE Cipher Processing:\n")
	fmt.Fprintf(file, "  Total time: %v\n", fpeTime)
	fmt.Fprintf(file, "  Rate: %.0f items/sec\n", float64(testCount)/fpeTime.Seconds())
	fmt.Fprintf(file, "  Average: %.3f μs per item\n", float64(fpeTime.Microseconds())/float64(testCount))

	// Performance comparison
	fastest := "Numbers"
	fastestTime := numbersTime
	if stringsTime < fastestTime {
		fastest = "Strings"
		fastestTime = stringsTime
	}
	if fpeTime < fastestTime {
		fastest = "FPE Cipher"
		fastestTime = fpeTime
	}

	fmt.Fprintf(file, "\n%s is the fastest\n", fastest)
	if fastestTime != numbersTime {
		fmt.Fprintf(file, "Numbers are %.1fx slower than %s\n", float64(numbersTime)/float64(fastestTime), fastest)
	}
	if fastestTime != stringsTime {
		fmt.Fprintf(file, "Strings are %.1fx slower than %s\n", float64(stringsTime)/float64(fastestTime), fastest)
	}
	if fastestTime != fpeTime {
		fmt.Fprintf(file, "FPE Cipher is %.1fx slower than %s\n", float64(fpeTime)/float64(fastestTime), fastest)
	}

	fmt.Fprintf(file, "\n=== DUPLICATE ANALYSIS ===\n")
	fmt.Fprintf(file, "Encrypted numbers: %d unique, %d duplicates (%.2f%%)\n",
		len(encryptedNumbersSet), duplicateNumbers, float64(duplicateNumbers)/float64(testCount)*100)
	fmt.Fprintf(file, "Encrypted strings: %d unique, %d duplicates (%.2f%%)\n",
		len(encryptedStringsSet), duplicateStrings, float64(duplicateStrings)/float64(testCount)*100)
	fmt.Fprintf(file, "Input-Output collisions: %d (%.2f%%)\n",
		inputOutputCollision, float64(inputOutputCollision)/float64(testCount)*100)

	fmt.Fprintf(file, "\n=== VERIFICATION ===\n")
	fmt.Fprintf(file, "Numbers test: %d/%d correct (%.2f%%)\n", correctNumbers, testCount, float64(correctNumbers)/float64(testCount)*100)
	fmt.Fprintf(file, "Strings test: %d/%d correct (%.2f%%)\n", correctStrings, testCount, float64(correctStrings)/float64(testCount)*100)

	// Print performance summary
	printPerformanceSummary(numbersTime, stringsTime, fpeTime, testCount)

	// Create and display ASCII charts
	numbersData := map[string]float64{
		"Numbers":    float64(numbersTime.Milliseconds()),
		"Strings":    float64(stringsTime.Milliseconds()),
		"FPE Cipher": float64(fpeTime.Milliseconds()),
	}

	fmt.Println(createASCIIChart("PERFORMANCE COMPARISON (ms)", numbersData, 40))

	// Accuracy chart
	accuracyData := map[string]float64{
		"Numbers":    float64(correctNumbers) / float64(testCount) * 100,
		"Strings":    float64(correctStrings) / float64(testCount) * 100,
		"FPE Cipher": float64(correctFPE) / float64(testCount) * 100,
	}

	fmt.Println(createASCIIChart("ACCURACY COMPARISON (%)", accuracyData, 40))

	fmt.Println("Benchmark completed!")
	fmt.Printf("Test data written to test.txt (%d lines total)\n", testCount*2)
	fmt.Println("Results written to out.txt")
	fmt.Printf("Numbers accuracy: %.2f%%, Strings accuracy: %.2f%%, FPE accuracy: %.2f%%\n",
		float64(correctNumbers)/float64(testCount)*100, float64(correctStrings)/float64(testCount)*100, float64(correctFPE)/float64(testCount)*100)
}
