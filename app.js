/**
 * Recipe Website – app.js
 * Uses localStorage to persist recipes as JSON.
 */

'use strict';

// ── Storage helpers ──────────────────────────────────────────────────────────

const STORAGE_KEY = 'recipeWebsiteData';

function loadData() {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (raw) return JSON.parse(raw);
  } catch (_) {}
  return { recipes: [] };
}

function saveData(data) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(data));
}

function getRecipes() {
  return loadData().recipes;
}

function saveRecipes(recipes) {
  saveData({ recipes });
}

function generateId() {
  return Date.now().toString(36) + Math.random().toString(36).slice(2);
}

// ── Seed data (runs only on first visit) ─────────────────────────────────────

function seedIfEmpty() {
  if (getRecipes().length > 0) return;
  const seed = [
    {
      id: generateId(),
      title: 'Classic Pancakes',
      description: 'Fluffy, golden pancakes perfect for a weekend breakfast.',
      categories: ['breakfast', 'vegetarian'],
      ingredients: ['1 cup all-purpose flour', '2 tbsp sugar', '1 tsp baking powder',
        '½ tsp salt', '1 cup milk', '1 egg', '2 tbsp melted butter'],
      steps: [
        'Whisk together flour, sugar, baking powder, and salt in a large bowl.',
        'In another bowl, beat milk, egg, and melted butter together.',
        'Pour wet ingredients into dry ingredients and stir until just combined (lumps are fine).',
        'Heat a non-stick pan over medium heat and lightly grease with butter.',
        'Pour ¼ cup of batter per pancake and cook until bubbles form (~2 min), then flip.',
        'Serve with maple syrup and fresh fruit.'],
      links: [
        { type: 'external', url: 'https://en.wikipedia.org/wiki/Pancake', label: 'About Pancakes' }
      ]
    },
    {
      id: generateId(),
      title: 'Simple Tomato Pasta',
      description: 'A quick weeknight pasta with a rich, garlicky tomato sauce.',
      categories: ['dinner', 'vegetarian'],
      ingredients: ['400 g spaghetti', '3 tbsp olive oil', '4 garlic cloves, minced',
        '1 can (400 g) crushed tomatoes', 'Salt and pepper', 'Fresh basil', 'Parmesan (optional)'],
      steps: [
        'Cook spaghetti according to package directions; reserve ½ cup pasta water.',
        'Heat olive oil in a large pan over medium heat. Add garlic and cook 1 minute.',
        'Add crushed tomatoes, season with salt and pepper, and simmer 10 minutes.',
        'Toss pasta with sauce, adding pasta water to loosen if needed.',
        'Serve topped with fresh basil and grated Parmesan.'],
      links: []
    },
    {
      id: generateId(),
      title: 'Avocado Toast',
      description: 'Creamy avocado on toasted bread, with toppings of your choice.',
      categories: ['breakfast', 'lunch', 'vegetarian'],
      ingredients: ['2 slices thick bread', '1 ripe avocado', 'Juice of ½ lemon',
        'Salt, pepper, chilli flakes', 'Optional: poached egg, cherry tomatoes'],
      steps: [
        'Toast bread until golden and crisp.',
        'Halve avocado, remove pit, and scoop flesh into a bowl.',
        'Mash avocado with lemon juice, salt, and pepper.',
        'Spread avocado mixture on toast.',
        'Top with chilli flakes and any optional toppings.'],
      links: []
    }
  ];
  saveRecipes(seed);
}

// ── State ────────────────────────────────────────────────────────────────────

let state = {
  activeCategory: null,   // null = all
  searchQuery: '',
  editingId: null,        // recipe id being edited, or null for new
  pendingDeleteId: null
};

// ── DOM refs ─────────────────────────────────────────────────────────────────

const $ = (sel, ctx = document) => ctx.querySelector(sel);
const $$ = (sel, ctx = document) => [...ctx.querySelectorAll(sel)];

const categoryList  = $('#category-list');
const recipeCards   = $('#recipe-cards');
const noRecipesMsg  = $('#no-recipes-msg');
const listHeading   = $('#list-heading');
const viewList      = $('#view-list');
const viewDetail    = $('#view-detail');
const recipeDetail  = $('#recipe-detail');
const searchInput   = $('#search-input');
const modalOverlay  = $('#modal-overlay');
const deleteOverlay = $('#delete-overlay');
const recipeForm    = $('#recipe-form');

// ── Category helpers ──────────────────────────────────────────────────────────

function getAllCategories(recipes) {
  const set = new Set();
  recipes.forEach(r => (r.categories || []).forEach(c => set.add(c.toLowerCase().trim())));
  return [...set].sort();
}

// ── Rendering ─────────────────────────────────────────────────────────────────

function renderCategories() {
  const recipes = getRecipes();
  const categories = getAllCategories(recipes);

  categoryList.innerHTML = '';

  // "All" entry
  const allLi = document.createElement('li');
  const allA  = document.createElement('a');
  allA.href = '#';
  allA.textContent = '🍽 All Recipes';
  allA.className = state.activeCategory === null ? 'active' : '';
  allA.addEventListener('click', e => { e.preventDefault(); selectCategory(null); });
  allLi.appendChild(allA);
  categoryList.appendChild(allLi);

  categories.forEach(cat => {
    const li = document.createElement('li');
    const a  = document.createElement('a');
    a.href = '#';
    a.textContent = capitalize(cat);
    a.className = state.activeCategory === cat ? 'active' : '';
    a.addEventListener('click', e => { e.preventDefault(); selectCategory(cat); });
    li.appendChild(a);
    categoryList.appendChild(li);
  });
}

function getFilteredRecipes() {
  let recipes = getRecipes();

  if (state.activeCategory) {
    recipes = recipes.filter(r =>
      (r.categories || []).map(c => c.toLowerCase().trim()).includes(state.activeCategory));
  }

  if (state.searchQuery) {
    const q = state.searchQuery.toLowerCase();
    recipes = recipes.filter(r =>
      r.title.toLowerCase().includes(q) ||
      (r.description || '').toLowerCase().includes(q) ||
      (r.categories || []).some(c => c.toLowerCase().includes(q)) ||
      (r.ingredients || []).some(i => i.toLowerCase().includes(q))
    );
  }

  return recipes;
}

function renderRecipeList() {
  const recipes = getFilteredRecipes();
  recipeCards.innerHTML = '';

  if (recipes.length === 0) {
    noRecipesMsg.classList.remove('hidden');
  } else {
    noRecipesMsg.classList.add('hidden');
    recipes.forEach(r => recipeCards.appendChild(buildCard(r)));
  }

  // heading
  if (state.searchQuery) {
    listHeading.textContent = `Search: "${state.searchQuery}"`;
  } else if (state.activeCategory) {
    listHeading.textContent = capitalize(state.activeCategory);
  } else {
    listHeading.textContent = 'All Recipes';
  }
}

function buildCard(recipe) {
  const card = document.createElement('div');
  card.className = 'recipe-card';
  card.setAttribute('data-id', recipe.id);

  const h3 = document.createElement('h3');
  h3.textContent = recipe.title;

  const p = document.createElement('p');
  p.textContent = recipe.description || '';

  const tags = document.createElement('div');
  tags.className = 'card-tags';
  (recipe.categories || []).forEach(c => {
    const span = document.createElement('span');
    span.className = 'tag';
    span.textContent = capitalize(c);
    tags.appendChild(span);
  });

  const actions = document.createElement('div');
  actions.className = 'card-actions';

  const editBtn = document.createElement('button');
  editBtn.className = 'btn-icon';
  editBtn.title = 'Edit recipe';
  editBtn.textContent = '✏️';
  editBtn.addEventListener('click', e => { e.stopPropagation(); openEditModal(recipe.id); });

  const delBtn = document.createElement('button');
  delBtn.className = 'btn-icon';
  delBtn.title = 'Delete recipe';
  delBtn.textContent = '🗑️';
  delBtn.addEventListener('click', e => { e.stopPropagation(); openDeleteModal(recipe.id); });

  actions.appendChild(editBtn);
  actions.appendChild(delBtn);

  card.appendChild(h3);
  card.appendChild(p);
  card.appendChild(tags);
  card.appendChild(actions);

  card.addEventListener('click', () => showDetail(recipe.id));

  return card;
}

function showDetail(id) {
  const recipes = getRecipes();
  const recipe = recipes.find(r => r.id === id);
  if (!recipe) return;

  recipeDetail.innerHTML = '';

  const h2 = document.createElement('h2');
  h2.textContent = recipe.title;

  // categories
  const meta = document.createElement('div');
  meta.className = 'detail-meta';
  (recipe.categories || []).forEach(c => {
    const span = document.createElement('span');
    span.className = 'tag';
    span.textContent = capitalize(c);
    meta.appendChild(span);
  });

  // actions
  const detailActions = document.createElement('div');
  detailActions.className = 'detail-actions';
  const editBtn = document.createElement('button');
  editBtn.className = 'btn btn-secondary';
  editBtn.textContent = '✏️ Edit';
  editBtn.addEventListener('click', () => openEditModal(recipe.id));
  const delBtn = document.createElement('button');
  delBtn.className = 'btn btn-danger';
  delBtn.textContent = '🗑️ Delete';
  delBtn.addEventListener('click', () => openDeleteModal(recipe.id));
  detailActions.appendChild(editBtn);
  detailActions.appendChild(delBtn);

  // description
  const descSection = document.createElement('section');
  const descP = document.createElement('p');
  descP.textContent = recipe.description || '';
  descSection.appendChild(descP);

  // ingredients
  const ingSection = document.createElement('section');
  const ingH3 = document.createElement('h3');
  ingH3.textContent = 'Ingredients';
  const ingUl = document.createElement('ul');
  (recipe.ingredients || []).forEach(ing => {
    const li = document.createElement('li');
    li.textContent = ing;
    ingUl.appendChild(li);
  });
  ingSection.appendChild(ingH3);
  ingSection.appendChild(ingUl);

  // steps
  const stepsSection = document.createElement('section');
  const stepsH3 = document.createElement('h3');
  stepsH3.textContent = 'Steps';
  const stepsOl = document.createElement('ol');
  (recipe.steps || []).forEach(step => {
    const li = document.createElement('li');
    li.textContent = step;
    stepsOl.appendChild(li);
  });
  stepsSection.appendChild(stepsH3);
  stepsSection.appendChild(stepsOl);

  recipeDetail.appendChild(h2);
  recipeDetail.appendChild(meta);
  recipeDetail.appendChild(detailActions);
  recipeDetail.appendChild(descSection);
  recipeDetail.appendChild(ingSection);
  recipeDetail.appendChild(stepsSection);

  // links
  const links = recipe.links || [];
  if (links.length > 0) {
    const linksSection = document.createElement('section');
    const linksH3 = document.createElement('h3');
    linksH3.textContent = 'Related Links';
    const linksDiv = document.createElement('div');
    linksDiv.className = 'detail-links';

    links.forEach(link => {
      const a = document.createElement('a');
      if (link.type === 'recipe') {
        const linked = recipes.find(r => r.id === link.id);
        a.textContent = '🔗 ' + (linked ? linked.title : 'Unknown recipe');
        a.href = '#';
        a.addEventListener('click', e => {
          e.preventDefault();
          if (linked) showDetail(linked.id);
        });
      } else {
        a.textContent = '🌐 ' + (link.label || link.url);
        a.href = link.url;
        a.target = '_blank';
        a.rel = 'noopener noreferrer';
      }
      linksDiv.appendChild(a);
    });

    linksSection.appendChild(linksH3);
    linksSection.appendChild(linksDiv);
    recipeDetail.appendChild(linksSection);
  }

  showView('detail');
}

// ── View switching ────────────────────────────────────────────────────────────

function showView(name) {
  viewList.classList.remove('active');
  viewList.classList.add('hidden');
  viewDetail.classList.remove('active');
  viewDetail.classList.add('hidden');

  if (name === 'list') {
    viewList.classList.add('active');
    viewList.classList.remove('hidden');
  } else {
    viewDetail.classList.add('active');
    viewDetail.classList.remove('hidden');
  }
}

function selectCategory(cat) {
  state.activeCategory = cat;
  state.searchQuery = '';
  searchInput.value = '';
  renderCategories();
  renderRecipeList();
  showView('list');
}

// ── Modal: Add / Edit ─────────────────────────────────────────────────────────

function openAddModal() {
  state.editingId = null;
  $('#modal-title').textContent = 'Add Recipe';
  recipeForm.reset();
  $('#form-id').value = '';
  clearDynamicList('ingredients-list');
  clearDynamicList('steps-list');
  clearDynamicList('links-list');
  addIngredientRow('');
  addStepRow('');
  modalOverlay.classList.remove('hidden');
  $('#form-title').focus();
}

function openEditModal(id) {
  const recipe = getRecipes().find(r => r.id === id);
  if (!recipe) return;
  state.editingId = id;
  $('#modal-title').textContent = 'Edit Recipe';
  $('#form-id').value = id;
  $('#form-title').value = recipe.title;
  $('#form-description').value = recipe.description || '';
  $('#form-categories').value = (recipe.categories || []).join(', ');

  clearDynamicList('ingredients-list');
  (recipe.ingredients || []).forEach(ing => addIngredientRow(ing));
  if ((recipe.ingredients || []).length === 0) addIngredientRow('');

  clearDynamicList('steps-list');
  (recipe.steps || []).forEach(step => addStepRow(step));
  if ((recipe.steps || []).length === 0) addStepRow('');

  clearDynamicList('links-list');
  (recipe.links || []).forEach(link => addLinkRow(link));

  modalOverlay.classList.remove('hidden');
  $('#form-title').focus();
}

function closeModal() {
  modalOverlay.classList.add('hidden');
  clearFormError();
}

function clearDynamicList(listId) {
  document.getElementById(listId).innerHTML = '';
}

function addIngredientRow(value) {
  const list = document.getElementById('ingredients-list');
  const row = buildTextRow(value, 'Ingredient', () => row.remove());
  list.appendChild(row);
}

function addStepRow(value) {
  const list = document.getElementById('steps-list');
  const row = buildTextRow(value, 'Step description', () => row.remove());
  list.appendChild(row);
}

function addLinkRow(link = {}) {
  const list = document.getElementById('links-list');
  const row = document.createElement('div');
  row.className = 'dynamic-item';

  const typeSelect = document.createElement('select');
  typeSelect.className = 'link-type';
  ['external', 'recipe'].forEach(t => {
    const opt = document.createElement('option');
    opt.value = t;
    opt.textContent = t === 'external' ? 'External URL' : 'Recipe';
    typeSelect.appendChild(opt);
  });
  typeSelect.value = link.type || 'external';

  const urlInput = document.createElement('input');
  urlInput.type = 'text';
  urlInput.className = 'link-url';
  urlInput.placeholder = 'URL or recipe ID';
  urlInput.value = link.url || link.id || '';

  const labelInput = document.createElement('input');
  labelInput.type = 'text';
  labelInput.className = 'link-label';
  labelInput.placeholder = 'Label (optional)';
  labelInput.value = link.label || '';

  typeSelect.addEventListener('change', () => {
    urlInput.placeholder = typeSelect.value === 'recipe' ? 'Recipe title search' : 'URL';
    labelInput.style.display = typeSelect.value === 'recipe' ? 'none' : '';
  });
  if (link.type === 'recipe') labelInput.style.display = 'none';

  const removeBtn = document.createElement('button');
  removeBtn.type = 'button';
  removeBtn.className = 'btn-icon';
  removeBtn.title = 'Remove';
  removeBtn.textContent = '✕';
  removeBtn.addEventListener('click', () => row.remove());

  row.appendChild(typeSelect);
  row.appendChild(urlInput);
  row.appendChild(labelInput);
  row.appendChild(removeBtn);
  list.appendChild(row);
}

function buildTextRow(value, placeholder, onRemove) {
  const row = document.createElement('div');
  row.className = 'dynamic-item';

  const input = document.createElement('input');
  input.type = 'text';
  input.value = value;
  input.placeholder = placeholder;

  const removeBtn = document.createElement('button');
  removeBtn.type = 'button';
  removeBtn.className = 'btn-icon';
  removeBtn.title = 'Remove';
  removeBtn.textContent = '✕';
  removeBtn.addEventListener('click', onRemove);

  row.appendChild(input);
  row.appendChild(removeBtn);
  return row;
}

function collectDynamicValues(listId) {
  return $$(`#${listId} .dynamic-item input[type="text"]`)
    .map(i => i.value.trim())
    .filter(Boolean);
}

function collectLinks() {
  return $$('#links-list .dynamic-item').map(row => {
    const type = row.querySelector('.link-type').value;
    const urlVal = row.querySelector('.link-url').value.trim();
    const labelVal = row.querySelector('.link-label')?.value ?? '';
    if (!urlVal) return null;

    if (type === 'recipe') {
      // resolve recipe by id or title
      const recipes = getRecipes();
      const match = recipes.find(r => r.id === urlVal ||
        r.title.toLowerCase() === urlVal.toLowerCase());
      return match ? { type: 'recipe', id: match.id } : null;
    }
    return { type: 'external', url: urlVal, label: labelVal.trim() };
  }).filter(Boolean);
}

function showFormError(msg) {
  const el = $('#form-error');
  el.textContent = msg;
  el.classList.remove('hidden');
  el.scrollIntoView({ block: 'nearest' });
}

function clearFormError() {
  const el = $('#form-error');
  el.textContent = '';
  el.classList.add('hidden');
}

function handleFormSubmit(e) {
  e.preventDefault();
  clearFormError();

  const title = $('#form-title').value.trim();
  if (!title) { showFormError('Please enter a recipe title.'); $('#form-title').focus(); return; }

  const categories = $('#form-categories').value
    .split(',').map(c => c.trim().toLowerCase()).filter(Boolean);
  if (categories.length === 0) { showFormError('Please enter at least one category.'); $('#form-categories').focus(); return; }

  const ingredients = collectDynamicValues('ingredients-list');
  if (ingredients.length === 0) { showFormError('Please add at least one ingredient.'); return; }

  const steps = collectDynamicValues('steps-list');
  if (steps.length === 0) { showFormError('Please add at least one step.'); return; }

  const links = collectLinks();

  const recipes = getRecipes();

  if (state.editingId) {
    const idx = recipes.findIndex(r => r.id === state.editingId);
    if (idx !== -1) {
      recipes[idx] = {
        ...recipes[idx],
        title,
        description: $('#form-description').value.trim(),
        categories,
        ingredients,
        steps,
        links
      };
    }
  } else {
    recipes.push({
      id: generateId(),
      title,
      description: $('#form-description').value.trim(),
      categories,
      ingredients,
      steps,
      links
    });
  }

  saveRecipes(recipes);
  closeModal();
  renderCategories();
  renderRecipeList();

  // If we edited the currently viewed detail, refresh it
  if (state.editingId && !viewDetail.classList.contains('hidden')) {
    showDetail(state.editingId);
  }
}

// ── Modal: Delete ─────────────────────────────────────────────────────────────

function openDeleteModal(id) {
  const recipe = getRecipes().find(r => r.id === id);
  if (!recipe) return;
  state.pendingDeleteId = id;
  $('#delete-recipe-name').textContent = recipe.title;
  deleteOverlay.classList.remove('hidden');
}

function closeDeleteModal() {
  state.pendingDeleteId = null;
  deleteOverlay.classList.add('hidden');
}

function confirmDelete() {
  if (!state.pendingDeleteId) return;
  const recipes = getRecipes().filter(r => r.id !== state.pendingDeleteId);
  saveRecipes(recipes);
  closeDeleteModal();
  renderCategories();
  renderRecipeList();
  showView('list');
}

// ── Utilities ─────────────────────────────────────────────────────────────────

function capitalize(str) {
  return str.charAt(0).toUpperCase() + str.slice(1);
}

// ── Event wiring ──────────────────────────────────────────────────────────────

function wireEvents() {
  $('#btn-add-recipe').addEventListener('click', openAddModal);
  $('#btn-cancel').addEventListener('click', closeModal);
  recipeForm.addEventListener('submit', handleFormSubmit);

  $('#btn-add-ingredient').addEventListener('click', () => addIngredientRow(''));
  $('#btn-add-step').addEventListener('click', () => addStepRow(''));
  $('#btn-add-link').addEventListener('click', () => addLinkRow());

  $('#btn-back').addEventListener('click', () => { showView('list'); });

  $('#btn-confirm-delete').addEventListener('click', confirmDelete);
  $('#btn-cancel-delete').addEventListener('click', closeDeleteModal);

  // Close modals on backdrop click
  modalOverlay.addEventListener('click', e => { if (e.target === modalOverlay) closeModal(); });
  deleteOverlay.addEventListener('click', e => { if (e.target === deleteOverlay) closeDeleteModal(); });

  // Escape key closes modals
  document.addEventListener('keydown', e => {
    if (e.key === 'Escape') { closeModal(); closeDeleteModal(); }
  });

  // Search
  searchInput.addEventListener('input', () => {
    state.searchQuery = searchInput.value.trim();
    state.activeCategory = null;
    renderCategories();
    renderRecipeList();
    showView('list');
  });
}

// ── Bootstrap ─────────────────────────────────────────────────────────────────

function init() {
  seedIfEmpty();
  wireEvents();
  renderCategories();
  renderRecipeList();
  showView('list');
}

document.addEventListener('DOMContentLoaded', init);
