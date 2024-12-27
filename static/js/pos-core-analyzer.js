const posAnalyzer = {
  get categories() {
    return Object.keys(COLORS);
  },

  analyze(text) {
    const doc = nlp(text.toLowerCase());
    const results = this.categories.reduce((acc, category) => {
      acc[category] = {
        words: [],
        frequencies: {}
      };
      return acc;
    }, {});

    const allTerms = doc.json();

    let hyphenatedWord = '';
    let hyphenatedTags = new Set();

    allTerms.forEach(sentence => {
      sentence.terms.forEach(term => {
        const word = term.text.toLowerCase().trim();
        const tags = term.tags || [];

        if (!word) return;

        const addToCategory = (category, word) => {
          if (!results[category].frequencies[word]) {
            results[category].frequencies[word] = 0;
            results[category].words.push(word);
          }
          results[category].frequencies[word]++;
        };

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
        }

        if (tags.includes('Pronoun') ||
          (tags.includes('Noun') && tags.includes('Possessive'))) {
          addToCategory('pronoun', word);
          return;
        }

        if (tags.includes('Verb') || tags.includes('Modal') ||
          tags.includes('Auxiliary') || tags.includes('Copula')) {
          addToCategory('verb', word);
        }

        if (tags.includes('Adjective')) {
          addToCategory('adverb', word);
        }

        if (tags.includes('Adverb') || tags.includes('Negative')) {
          addToCategory('adverb', word);
        }

        if (tags.includes('Preposition')) {
          addToCategory('preposition', word);
        }

        if (tags.includes('Conjunction')) {
          addToCategory('conjunction', word);
        }

        if (tags.includes('Determiner') || tags.includes('Article') ||
          tags.some(tag => ['Value', 'Cardinal', 'TextValue'].includes(tag))) {
          addToCategory('determiner', word);
        }

        if (tags.includes('Expression') || tags.includes('Interjection')) {
          addToCategory('interjection', word);
        }

        if ((tags.includes('Noun') || tags.includes('Singular') ||
          tags.includes('Plural') || tags.includes('Uncountable')) &&
          !tags.includes('Pronoun')) {
          addToCategory('noun', word);
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

    return finalResults;
  }
};

window.posAnalyzer = posAnalyzer;