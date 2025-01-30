const posAnalyzer = {
  get categories() {
    return Object.keys(COLORS);
  },

  analyze(text, ignoredWords = '') {
    const doc = nlp(text.toLowerCase());
    const results = this.categories.reduce((acc, category) => {
      acc[category] = { words: [], frequencies: {} };
      return acc;
    }, {});

    const ignoredWordsSet = new Set(
      ignoredWords
        .split(',')
        .map(w => w.trim().toLowerCase())
        .filter(Boolean)
    );

    const processedWords = new Set();
    const allWords = new Set();

    const allTerms = doc.json();
    let hyphenatedWord = '';
    let hyphenatedTags = new Set();

    allTerms.forEach(sentence => {
      sentence.terms.forEach(term => {
        const word = term.text.toLowerCase().trim();
        const tags = term.tags || [];

        if (!word || ignoredWordsSet.has(word)) return;
        allWords.add(word);

        const addToCategory = (category, word) => {
          if (!results[category].frequencies[word]) {
            results[category].frequencies[word] = 0;
            results[category].words.push(word);
          }
          results[category].frequencies[word]++;
          processedWords.add(word);
        };

        if (tags.includes('Imperative') || tags.includes('Infinitive') || tags.includes('PresentTense')) {
          addToCategory('verb', word);
          return;
        }

        if (tags.includes('Date')) {
          addToCategory('noun', word);
          return;
        }

        if (tags.includes('There')) {
          addToCategory('adverb', word);
          return;
        }

        if (tags.includes('QuestionWord')) {
          if (['who', 'whom', 'whose', 'what'].includes(word)) {
            addToCategory('pronoun', word);
          } else {
            addToCategory('adverb', word);
          }
          return;
        }

        if (word === 'to') {
          addToCategory('preposition', word);
          return;
        }

        if (tags.includes('Hyphenated')) {
          if (hyphenatedWord) {
            hyphenatedWord += '-' + word;
            tags.forEach(tag => hyphenatedTags.add(tag));
          } else {
            hyphenatedWord = word;
            tags.forEach(tag => hyphenatedTags.add(tag));
          }
          return;
        } else if (hyphenatedWord) {
          const fullWord = hyphenatedWord;
          const tagArray = Array.from(hyphenatedTags);

          if (tagArray.includes('Adjective')) {
            addToCategory('adjective', fullWord);
          } else if (tagArray.includes('Verb')) {
            addToCategory('verb', fullWord);
          } else if (tagArray.includes('Noun')) {
            addToCategory('noun', fullWord);
          } else if (tagArray.includes('Adverb') || tagArray.includes('Negative')) {
            addToCategory('adverb', fullWord);
          }

          hyphenatedWord = '';
          hyphenatedTags.clear();
          return;
        }

        if (tags.includes('Determiner') || tags.includes('Article')) {
          addToCategory('determiner', word);
          return;
        }

        if (tags.includes('Pronoun') || (tags.includes('Noun') && tags.includes('Possessive'))) {
          addToCategory('pronoun', word);
          return;
        }

        if (tags.includes('Verb') || tags.includes('Modal') || tags.includes('Auxiliary') || tags.includes('Copula')) {
          addToCategory('verb', word);
          return;
        }

        if (tags.includes('Adjective')) {
          addToCategory('adjective', word);
          return;
        }

        if (tags.includes('Adverb') || tags.includes('Negative')) {
          addToCategory('adverb', word);
          return;
        }

        if (tags.includes('Preposition')) {
          addToCategory('preposition', word);
          return;
        }

        if (tags.includes('Conjunction')) {
          addToCategory('conjunction', word);
          return;
        }

        if (tags.includes('Value') || tags.includes('Cardinal')) {
          addToCategory('determiner', word);
          return;
        }

        if ((tags.includes('Noun') || tags.includes('Singular') || tags.includes('Plural') || tags.includes('Uncountable')) && !tags.includes('Pronoun')) {
          addToCategory('noun', word);
          return;
        }

        if (tags.includes('Expression') || tags.includes('Interjection')) {
          addToCategory('interjection', word);
          return;
        }
      });
    });

    const finalResults = {};
    Object.entries(results).forEach(([category, data]) => {
      finalResults[category] = {
        unique: data.words,
        total: Object.values(data.frequencies).reduce((sum, count) => sum + count, 0)
      };
    });

    finalResults._debug = {
      missedWords: Array.from(allWords).filter(word =>
        !processedWords.has(word) &&
        !ignoredWordsSet.has(word)
      ).sort(),
      allWords: Array.from(allWords).sort(),
      processedWords: Array.from(processedWords).sort(),
      rawTerms: allTerms.map(sentence => ({
        text: sentence.text,
        terms: sentence.terms.map(term => ({
          text: term.text,
          tags: term.tags
        }))
      }))
    };

    return finalResults;
  }
};

window.posAnalyzer = posAnalyzer;