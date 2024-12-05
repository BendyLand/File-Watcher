package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

const prevFilePath = "watcher/prev.json"
const changedFilesPath = "watcher/changed_files.txt"

// FileHashes stores filenames and their hashes.
type FileHashes map[string]string

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: watcher <directory_path>")
		os.Exit(1)
	}
	if os.Args[1] == "init" {
		err := initWatcher()
		if err != nil {
			fmt.Printf("Error initializing watcher structure: %s\n", err)
			os.Exit(1)
		}
		fmt.Println("Watcher initialized successfully!")
		return
	} else if os.Args[1] == "help" {
		printHelpMenu()
		return
	} else if os.Args[1] == "clear" {
		err := clearPrevJson()
		if err != nil {
			fmt.Printf("Error cleaning 'prev.json': %s\n", err)
			os.Exit(1)
		}
		fmt.Println("'prev.json' cleared successfully!")
		return
	}
	path := os.Args[1]
	prevHashes := loadPrevHashes()
	currHashes := computeHashes(path)
	latestChanges := make(FileHashes)
	for file, _ := range currHashes {
		if _, ok := prevHashes[file]; ok || len(prevHashes) == 0 {
			if prevHashes[file] != currHashes[file] {
				latestChanges[file] = currHashes[file]
			}
		} else {
			latestChanges[file] = currHashes[file]
		}
	}
	err := saveHashes(currHashes, latestChanges)
	if err != nil {
		// No changes
		writeTxtFile(latestChanges) // need to make sure the txt file is empty
		fmt.Println(err)
	} else {
		fmt.Println("Changed files:")
		for file, _ := range latestChanges {
			fmt.Println(file)
		}
		fmt.Println("\nFiles written to 'watcher/changed_files.txt'.")
	}
}

// computeHashes computes SHA-256 hashes for the files in the given directory.
func computeHashes(dir string) FileHashes {
	hashes := make(FileHashes)
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Fatalf("Error reading directory: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		hash, err := hashFile(path)
		if err != nil {
			log.Printf("Error hashing file %s: %v", path, err)
			continue
		}
		hashes[path] = hash
	}
	return hashes
}

// hashFile computes the SHA-256 hash of a file's contents.
func hashFile(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("error opening file %s: %w", filename, err)
	}
	defer f.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return "", fmt.Errorf("error hashing file %s: %w", filename, err)
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func loadPrevHashes() FileHashes {
	data, err := os.ReadFile(prevFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(FileHashes) // If the file doesn't exist, return an empty map.
		}
		log.Fatalf("Error reading hash file: %v", err)
	}
	var hashes FileHashes
	if err := json.Unmarshal(data, &hashes); err != nil {
		log.Fatalf("Error unmarshaling hash file: %v", err)
	}
	return hashes
}

func saveHashes(currHashes, changedHashes FileHashes) error {
	if len(changedHashes) == 0 {
		// Avoid resetting the JSON if no changes occurred.
		return fmt.Errorf("No changes detected. 'changed_files.txt' cleared.")
	}
	currData, err := json.MarshalIndent(currHashes, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling current hashes: %v", err)
	}
	if err := os.WriteFile(prevFilePath, currData, 0644); err != nil {
		fmt.Printf("Error: %s.\nPlease run `watcher init` to generate necessary files.\n", err)
		os.Exit(1)
	}
	writeTxtFile(changedHashes)
	return nil
}

func writeTxtFile(hashes FileHashes) {
	contents := ""
	for file, _ := range hashes {
		contents += file + "\n"
	}
	if err := os.WriteFile(changedFilesPath, []byte(contents), 0644); err != nil {
		log.Fatalf("Error writing text file: %v\n", err)
	}
}

func initWatcher() error {
	os.Mkdir("watcher", 0755)
	f, err := os.OpenFile("watcher/prev.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	_, err = f.Write([]byte("{}"))
	if err != nil {
		return err
	}
	f.Close()
	f2, err := os.OpenFile("watcher/changed_files.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	_, err = f2.Write([]byte(""))
	if err != nil {
		return err
	}
	f2.Close()
	return nil
}

func printHelpMenu() {
	fmt.Println(
		"Welcome to the file watcher help menu!\n\n" +
		"Usage: watcher <directory_path> <opt_command>\n\n" +
		"Valid commands:\n" +
		"help  - Shows this menu.\n" + 
		"init  - Generate the necessary directory structure for the tool.\n" +
		"clear - Clears 'prev.json' in case it gets corrupted.\n" + 
		"        Running the tool again will repopulate it.",
	)
}

func clearPrevJson() error {
	err := os.WriteFile("watcher/prev.json", []byte("{}"), 0644)
	if err != nil {
		return err
	}
	return nil
}
