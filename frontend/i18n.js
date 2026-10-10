'use strict';

// Minimal translation layer. Code and markup are written in English, which is
// also the fallback. Set LANG to a key of TRANSLATIONS to translate the UI.
// Recipe content (titles, ingredients, ...) is user data and is never translated.
const LANG = 'sv';

const TRANSLATIONS = {
  sv: {
    ui: {
      'Admin login': 'Admin-inloggning',
      'Log in': 'Logga in',
      'Log out': 'Logga ut',
      'PIN code': 'PIN-kod',
      'Invalid PIN': 'Fel PIN-kod',
      'Login failed': 'Inloggningen misslyckades',
      'Toggle menu': 'Växla meny',
      'Menu': 'Meny',
      'Add Recipe': 'Lägg till recept',
      'Search recipes': 'Sök recept',
      'Search': 'Sök',
      'Tag': 'Tagg',
      'No categories yet. Add a recipe to get started!': 'Inga kategorier än. Lägg till ett recept för att komma igång!',
      'Categories': 'Kategorier',
      'All recipes': 'Alla recept',
      'No recipes found.': 'Inga recept hittades.',
      'Back': 'Tillbaka',
      'Edit Recipe': 'Redigera recept',
      'Edit recipe': 'Redigera recept',
      'Delete Recipe': 'Ta bort recept',
      'Delete recipe': 'Ta bort recept',
      'Title': 'Titel',
      'Recipe title': 'Receptets titel',
      'Description': 'Beskrivning',
      'A short description of the recipe': 'En kort beskrivning av receptet',
      'Oven Temperature': 'Ugnstemperatur',
      'Oven mode': 'Ugnsläge',
      'Fan oven': 'Varmluft',
      'Conventional (top and bottom heat)': 'Över- och undervärme',
      'Please select an oven mode.': 'Välj ugnsläge.',
      '(optional)': '(valfritt)',
      'from °F': 'från °F',
      '(comma-separated)': '(kommaseparerade)',
      'e.g. Breakfast, Vegetarian': 't.ex. Frukost, Vegetariskt',
      'Tags': 'Taggar',
      '(comma-separated, searchable)': '(kommaseparerade, sökbara)',
      'e.g. quick, spicy, comfort-food': 't.ex. snabbt, starkt, husmanskost',
      'Ingredients': 'Ingredienser',
      'Ingredient': 'Ingrediens',
      'Add Ingredient': 'Lägg till ingrediens',
      'Steps': 'Instruktioner',
      'Step': 'Steg',
      'Step description': 'Beskriv steget',
      'Add Step': 'Lägg till steg',
      'Links': 'Länkar',
      'Related Links': 'Relaterade länkar',
      'Add Link': 'Lägg till länk',
      'Photo': 'Foto',
      '(optional, max 5 MB)': '(valfritt, max 5 MB)',
      'Current photo': 'Nuvarande foto',
      'Remove photo': 'Ta bort foto',
      'Drag the photo to choose how it appears in the recipe list': 'Dra i bilden för att välja hur den visas i receptlistan',
      'Zoom': 'Zoom',
      'Save Recipe': 'Spara recept',
      'Cancel': 'Avbryt',
      'View full image': 'Visa hela bilden',
      'Delete': 'Ta bort',
      'Are you sure you want to delete': 'Är du säker på att du vill ta bort',
      '? This cannot be undone.': '? Det går inte att ångra.',
      'recipe': 'recept',
      'recipes': 'recept',
      'Recipe': 'Recept',
      'Unknown recipe': 'Okänt recept',
      'External URL': 'Extern URL',
      'Type recipe name…': 'Skriv receptnamn…',
      'Label (optional)': 'Etikett (valfritt)',
      'Remove': 'Ta bort',
      'Drag to reorder': 'Dra för att ändra ordning',
      'Please enter a recipe title.': 'Ange en titel för receptet.',
      'Please enter at least one category.': 'Ange minst en kategori.',
      'Please add at least one ingredient.': 'Lägg till minst en ingrediens.',
      'Please add at least one step.': 'Lägg till minst ett steg.',
      'Failed to save recipe. Please try again.': 'Kunde inte spara receptet. Försök igen.',
      'Cannot delete: linked by': 'Kan inte tas bort: länkas från',
      'Failed to delete recipe.': 'Kunde inte ta bort receptet.',
      'Request failed': 'Förfrågan misslyckades',
      'Network error': 'Nätverksfel',
      'Unexpected server response': 'Oväntat svar från servern',
      'Upload failed': 'Uppladdningen misslyckades',
    },
    // Keyed by the stored (lowercase English) category name.
    categories: {
      breakfast: 'Frukost', lunch: 'Lunch', dinner: 'Middag', vegetarian: 'Vegetariskt',
      vegan: 'Veganskt', dessert: 'Efterrätt', snack: 'Snacks', soup: 'Soppa',
      pasta: 'Pasta', pizza: 'Pizza', salad: 'Sallad', meat: 'Kött',
      fish: 'Fisk', seafood: 'Skaldjur', baking: 'Bakning', bread: 'Bröd',
      drinks: 'Drycker', cocktail: 'Cocktail',
    },
  },
};

const _dict = TRANSLATIONS[LANG] || { ui: {}, categories: {} };

function t(text) {
  return _dict.ui[text] || text;
}

// Translated display name for a stored category, or '' if there is none.
function tCategory(cat) {
  return _dict.categories[String(cat).toLowerCase().trim()] || '';
}

// Translates the static markup: text nodes and common text attributes whose
// trimmed English value has a translation. Whitespace around text is kept.
function translateDOM(root = document.body) {
  if (!TRANSLATIONS[LANG]) return;
  document.documentElement.lang = LANG;

  const walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT);
  for (let node = walker.nextNode(); node; node = walker.nextNode()) {
    const text = node.nodeValue.trim();
    if (text && _dict.ui[text]) node.nodeValue = node.nodeValue.replace(text, _dict.ui[text]);
  }

  root.querySelectorAll('[placeholder],[title],[aria-label],[alt]').forEach(el => {
    ['placeholder', 'title', 'aria-label', 'alt'].forEach(attr => {
      const v = el.getAttribute(attr);
      if (v && _dict.ui[v]) el.setAttribute(attr, _dict.ui[v]);
    });
  });
}
