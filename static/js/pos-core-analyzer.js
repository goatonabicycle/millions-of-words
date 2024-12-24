const posAnalyzer = {
  get categories() {
    return Object.keys(COLORS);
  },

  analyze(text) {
    const doc = nlp(text.toLowerCase());
    const results = this.categories.reduce((acc, category) => {
      acc[category] = [];
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

        if (tags.includes('Date')) {
          if (!results.noun.includes(word)) {
            results.noun.push(word);
          }
          return;
        }

        if (tags.includes('There')) {
          if (!results.adverb.includes(word)) {
            results.adverb.push(word);
          }
          return;
        }

        if (tags.includes('QuestionWord')) {
          if (['who', 'whom', 'whose', 'what'].includes(word)) {
            if (!results.pronoun.includes(word)) {
              results.pronoun.push(word);
            }
          } else {
            if (!results.adverb.includes(word)) {
              results.adverb.push(word);
            }
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
            if (!results.adjective.includes(fullWord)) {
              results.adjective.push(fullWord);
            }
          } else if (tagArray.includes('Verb')) {
            if (!results.verb.includes(fullWord)) {
              results.verb.push(fullWord);
            }
          } else if (tagArray.includes('Noun')) {
            if (!results.noun.includes(fullWord)) {
              results.noun.push(fullWord);
            }
          } else if (tagArray.includes('Adverb') || tagArray.includes('Negative')) {
            if (!results.adverb.includes(fullWord)) {
              results.adverb.push(fullWord);
            }
          }

          hyphenatedWord = '';
          hyphenatedTags.clear();
        }

        if (tags.includes('Verb') || tags.includes('Modal') ||
          tags.includes('Auxiliary') || tags.includes('Copula')) {
          if (!results.verb.includes(word)) {
            results.verb.push(word);
          }
        }

        if (tags.includes('Adjective')) {
          if (!results.adjective.includes(word)) {
            results.adjective.push(word);
          }
        }

        if (tags.includes('Adverb') || tags.includes('Negative')) {
          if (!results.adverb.includes(word)) {
            results.adverb.push(word);
          }
        }

        if (tags.includes('Pronoun')) {
          if (!results.pronoun.includes(word)) {
            results.pronoun.push(word);
          }
        }

        if (tags.includes('Preposition')) {
          if (!results.preposition.includes(word)) {
            results.preposition.push(word);
          }
        }

        if (tags.includes('Conjunction')) {
          if (!results.conjunction.includes(word)) {
            results.conjunction.push(word);
          }
        }

        if (tags.includes('Determiner') || tags.includes('Article') ||
          tags.some(tag => ['Value', 'Cardinal', 'TextValue'].includes(tag))) {
          if (!results.determiner.includes(word)) {
            results.determiner.push(word);
          }
        }

        if (tags.includes('Expression') || tags.includes('Interjection')) {
          if (!results.interjection.includes(word)) {
            results.interjection.push(word);
          }
        }

        if ((tags.includes('Noun') || tags.includes('Singular') ||
          tags.includes('Plural') || tags.includes('Uncountable')) &&
          !tags.includes('Pronoun')) {
          if (!results.noun.includes(word)) {
            results.noun.push(word);
          }
        }
      });
    });

    return results;
  }
};

window.posAnalyzer = posAnalyzer;