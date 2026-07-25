function toggleRecipe() {
  const type = document.getElementById('entry_type').value;
  document.getElementById('recipe-fields').style.display = (type === 'recipe') ? 'block' : 'none';
}

function makeRemoveButton() {
  const btn = document.createElement('button');
  btn.type = 'button';
  btn.className = 'secondary remove-row';
  btn.textContent = 'x';
  return btn;
}

function addIngredientRow() {
  const row = document.createElement('div');
  row.className = 'repeat-row';

  const amount = document.createElement('input');
  amount.type = 'text';
  amount.name = 'ingredient_amount';
  amount.placeholder = 'Quantite (ex: 200 g)';

  const item = document.createElement('input');
  item.type = 'text';
  item.name = 'ingredient_item';
  item.placeholder = 'Ingredient';

  row.appendChild(amount);
  row.appendChild(item);
  row.appendChild(makeRemoveButton());
  document.getElementById('ingredients-rows').appendChild(row);
}

function addStepRow() {
  const row = document.createElement('div');
  row.className = 'repeat-row';

  const step = document.createElement('input');
  step.type = 'text';
  step.name = 'step_text';
  step.placeholder = "Decrire l'etape";

  row.appendChild(step);
  row.appendChild(makeRemoveButton());
  document.getElementById('steps-rows').appendChild(row);
}

document.getElementById('entry_type').addEventListener('change', toggleRecipe);
document.getElementById('add-ingredient').addEventListener('click', addIngredientRow);
document.getElementById('add-step').addEventListener('click', addStepRow);

// Delegation : fonctionne aussi pour les lignes ajoutees dynamiquement.
document.addEventListener('click', function (e) {
  if (e.target.classList.contains('remove-row')) {
    e.target.parentElement.remove();
  }
});

toggleRecipe();
