# Simple-Dictionary
A simple dictionary server written in Go.

### What is this for?
I often read books and I have to write down a lot of new words. My notes are already full and I don’t like the format of writing in them at all. So I thought, why not write an online dictionary that will store the words that I have read?

### List of endpoints:

1. Add a new word to the dictionary.

    ```
    GET /add
    ```
  
2. Remove a word from the dictionary.

    ```
    GET /delete
    ```
    
3. View the entire dictionary.

    ```
    GET /get
    ```

4. Look up words from the dictionary starting with the specified letter.

    ```
    GET /get-by-letter
    ```
    
