const posAnalyzer = {
  get categories() {
    return Object.keys(COLORS);
  },

  questionWords: {
    pronouns: ['who', 'whom', 'whose', 'which', 'what', "what's", 'that', 'whatever', 'whichever', 'whomever', 'whatsoever'],
    adverbs: ['why', 'how', 'when', 'where', 'whenever', 'wherever', 'however', 'whereby', 'wherefore', 'whether']
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

    allTerms.forEach(sentence => {
      let hyphenatedParts = [];
      let hyphenatedTags = new Set();

      sentence.terms.forEach((term, index) => {
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


        if (tags.includes('Hyphenated')) {
          hyphenatedParts.push(word);
          tags.forEach(tag => hyphenatedTags.add(tag));

          const nextTerm = sentence.terms[index + 1];
          if (!nextTerm || !nextTerm.tags.includes('Hyphenated')) {
            const fullWord = hyphenatedParts.join('-');
            const tagArray = Array.from(hyphenatedTags);

            if (tagArray.includes('Verb')) {
              addToCategory('verb', fullWord);
            } else if (tagArray.includes('Adjective')) {
              addToCategory('adjective', fullWord);
            } else {
              addToCategory('noun', fullWord);
            }

            hyphenatedParts = [];
            hyphenatedTags.clear();
          }
          return;
        }

        if (tags.includes('Value') || tags.includes('Cardinal') || tags.includes('TextValue')) {
          addToCategory('determiner', word);
          return;
        }

        if (tags.includes('Date')) {
          addToCategory('noun', word);
          return;
        }

        if (tags.includes('Expression') || tags.includes('Negative')) {
          addToCategory('interjection', word);
          return;
        }

        if (tags.includes('There')) {
          addToCategory('adverb', word);
          return;
        }

        if (tags.includes('QuestionWord')) {
          if (this.questionWords.pronouns.includes(word)) {
            addToCategory('pronoun', word);
            return;
          } else if (this.questionWords.adverbs.includes(word)) {
            addToCategory('adverb', word);
            return;
          }
        }

        if (tags.some(tag => [
          'Verb', 'Infinitive', 'Gerund', 'PastTense', 'PresentTense',
          'Modal', 'Auxiliary', 'Copula'
        ].includes(tag))) {
          if (!tags.includes('Possessive')) {
            addToCategory('verb', word);
            return;
          }
        }

        if (tags.includes('Adjective')) {
          addToCategory('adjective', word);
          return;
        }

        if (tags.includes('Adverb')) {
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

        if (tags.includes('Determiner') || tags.includes('Article')) {
          addToCategory('determiner', word);
          return;
        }

        if (tags.includes('Pronoun') || (tags.includes('Noun') && tags.includes('Possessive'))) {
          addToCategory('pronoun', word);
          return;
        }

        if (tags.includes('Expression') || tags.includes('Interjection')) {
          addToCategory('interjection', word);
          return;
        }

        if (tags.includes('Noun') || tags.includes('Singular') ||
          tags.includes('Plural') || tags.includes('Uncountable')) {
          addToCategory('noun', word);
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
      missedWords: Array.from(allWords)
        .filter(word => !processedWords.has(word) && !ignoredWordsSet.has(word))
        .sort(),
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