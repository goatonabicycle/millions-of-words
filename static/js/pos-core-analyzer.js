
const posAnalyzer = {
  TRADITIONAL_MAPPINGS: {
    INTERROGATIVE_PRONOUNS: ['who', 'which', 'what', "what's", 'whom', 'whose'],
    INTERROGATIVE_ADVERBS: ['when', 'where', 'why', 'how'],
    EXISTENTIAL_ADVERBS: ['there', "there's"]
  },

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

    allTerms.forEach(sentence => {
      sentence.terms.forEach(term => {
        const word = term.text.toLowerCase();
        const tags = term.tags || [];
        let categorized = false;

        if (this.TRADITIONAL_MAPPINGS.INTERROGATIVE_PRONOUNS.includes(word)) {
          results.pronoun.push(word);
          categorized = true;
        } else if (this.TRADITIONAL_MAPPINGS.INTERROGATIVE_ADVERBS.includes(word) ||
          this.TRADITIONAL_MAPPINGS.EXISTENTIAL_ADVERBS.includes(word)) {
          results.adverb.push(word);
          categorized = true;
        }

        if (!categorized) {
          if (word.includes('-')) {
            if (tags.includes('Noun')) results.noun.push(word);
            if (tags.includes('Adjective')) results.adjective.push(word);
            if (tags.includes('Verb')) results.verb.push(word);
            categorized = true;
          }

          if (tags.some(tag => ['Value', 'Cardinal', 'TextValue', 'Date', 'Duration'].includes(tag))) {
            results.determiner.push(word);
            categorized = true;
          }

          if (tags.includes('Pronoun') || (tags.includes('Possessive') && !tags.includes('Noun'))) {
            results.pronoun.push(word);
            categorized = true;
          }

          if (tags.includes('Preposition')) {
            results.preposition.push(word);
            categorized = true;
          }

          if (tags.includes('Conjunction')) {
            results.conjunction.push(word);
            categorized = true;
          }

          if (tags.includes('Determiner') || word === 'the' || word === 'a' || word === 'an') {
            results.determiner.push(word);
            categorized = true;
          }

          if (tags.includes('Verb') || tags.includes('Infinitive') ||
            tags.includes('PastTense') || tags.includes('Copula')) {
            results.verb.push(word);
            categorized = true;
          }

          if (tags.includes('Adverb') || tags.includes('Negative')) {
            results.adverb.push(word);
            categorized = true;
          }

          if (tags.includes('Adjective')) {
            results.adjective.push(word);
            categorized = true;
          }

          if (tags.includes('Expression') || tags.includes('Interjection')) {
            results.interjection.push(word);
            categorized = true;
          }

          if (tags.includes('Noun') || tags.includes('Singular') || tags.includes('Plural')) {
            results.noun.push(word);
            categorized = true;
          }
        }
      });
    });

    return results;
  }
};

window.posAnalyzer = posAnalyzer;