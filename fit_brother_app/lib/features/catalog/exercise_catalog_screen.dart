import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../auth/auth_provider.dart';
import 'catalog_provider.dart';

class ExerciseCatalogScreen extends ConsumerStatefulWidget {
  const ExerciseCatalogScreen({super.key});

  @override
  ConsumerState<ExerciseCatalogScreen> createState() =>
      _ExerciseCatalogScreenState();
}

class _ExerciseCatalogScreenState extends ConsumerState<ExerciseCatalogScreen> {
  final _searchCtrl = TextEditingController();
  CatalogSearchQuery _query = const CatalogSearchQuery();

  @override
  void dispose() {
    _searchCtrl.dispose();
    super.dispose();
  }

  void _applySearch() {
    setState(() {
      _query = _query.copyWith(query: _searchCtrl.text.trim(), offset: 0);
    });
  }

  void _clearFilters() {
    _searchCtrl.clear();
    setState(() {
      _query = const CatalogSearchQuery();
    });
  }

  @override
  Widget build(BuildContext context) {
    final isTrainer = ref.watch(authProvider.select((s) => s.role)) == 'TRAINER';
    final filtersAsync = ref.watch(catalogFiltersProvider);
    final searchAsync = ref.watch(catalogSearchProvider(_query));

    return Scaffold(
      backgroundColor: const Color(0xFFF5F7FA),
      appBar: AppBar(
        title: const Text('Biblioteca de Exercícios'),
        backgroundColor: Colors.white,
        foregroundColor: Colors.black87,
        elevation: 1,
      ),
      floatingActionButton: isTrainer
          ? FloatingActionButton.extended(
              onPressed: () => context.push('/catalog/create'),
              icon: const Icon(Icons.add),
              label: const Text('Novo Exercício'),
            )
          : null,
      body: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            _buildSearchBar(),
            const SizedBox(height: 12),
            _buildScopeChips(),
            const SizedBox(height: 12),
            filtersAsync.when(
              loading: () => const SizedBox.shrink(),
              error: (_, _) => const SizedBox.shrink(),
              data: (filters) => _buildDropdownFilters(filters),
            ),
            const SizedBox(height: 12),
            Expanded(
              child: searchAsync.when(
                loading: () => const Center(child: CircularProgressIndicator()),
                error: (err, _) => Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      const Icon(Icons.error_outline, color: Colors.red, size: 42),
                      const SizedBox(height: 8),
                      Text(
                        'Erro ao carregar biblioteca:\n$err',
                        textAlign: TextAlign.center,
                        style: const TextStyle(color: Colors.red),
                      ),
                      const SizedBox(height: 12),
                      ElevatedButton(
                        onPressed: () => ref.invalidate(catalogSearchProvider(_query)),
                        child: const Text('Tentar novamente'),
                      ),
                    ],
                  ),
                ),
                data: (res) {
                  if (res.items.isEmpty) {
                    return Center(
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          const Icon(Icons.fitness_center, size: 56, color: Colors.grey),
                          const SizedBox(height: 12),
                          const Text(
                            'Nenhum exercício encontrado com os filtros selecionados.',
                            textAlign: TextAlign.center,
                            style: TextStyle(color: Colors.black54),
                          ),
                          const SizedBox(height: 12),
                          TextButton(
                            onPressed: _clearFilters,
                            child: const Text('Limpar filtros'),
                          ),
                        ],
                      ),
                    );
                  }

                  return RefreshIndicator(
                    onRefresh: () async => ref.invalidate(catalogSearchProvider(_query)),
                    child: ListView.separated(
                      itemCount: res.items.length,
                      separatorBuilder: (_, _) => const SizedBox(height: 8),
                      itemBuilder: (context, index) =>
                          _buildExerciseCard(context, res.items[index]),
                    ),
                  );
                },
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildSearchBar() {
    return Row(
      children: [
        Expanded(
          child: TextField(
            controller: _searchCtrl,
            decoration: InputDecoration(
              hintText: 'Buscar exercício em português...',
              prefixIcon: const Icon(Icons.search),
              suffixIcon: _searchCtrl.text.isNotEmpty
                  ? IconButton(
                      icon: const Icon(Icons.clear),
                      onPressed: () {
                        _searchCtrl.clear();
                        _applySearch();
                      },
                    )
                  : null,
              filled: true,
              fillColor: Colors.white,
              border: OutlineInputBorder(
                borderRadius: BorderRadius.circular(10),
              ),
              contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
            ),
            textInputAction: TextInputAction.search,
            onSubmitted: (_) => _applySearch(),
          ),
        ),
        const SizedBox(width: 8),
        ElevatedButton(
          onPressed: _applySearch,
          child: const Text('Buscar'),
        ),
      ],
    );
  }

  Widget _buildScopeChips() {
    return Wrap(
      spacing: 8,
      runSpacing: 8,
      children: [
        ChoiceChip(
          label: const Text('Todos'),
          selected: _query.scope == 'all',
          onSelected: (_) => setState(() => _query = _query.copyWith(scope: 'all', offset: 0)),
        ),
        ChoiceChip(
          label: const Text('Catálogo Global'),
          selected: _query.scope == 'global',
          onSelected: (_) => setState(() => _query = _query.copyWith(scope: 'global', offset: 0)),
        ),
        ChoiceChip(
          label: const Text('Meus Customizados'),
          selected: _query.scope == 'custom',
          onSelected: (_) => setState(() => _query = _query.copyWith(scope: 'custom', offset: 0)),
        ),
      ],
    );
  }

  Widget _buildDropdownFilters(CatalogFiltersData filters) {
    return Wrap(
      spacing: 8,
      runSpacing: 8,
      children: [
        SizedBox(
          width: 190,
          child: DropdownButtonFormField<String>(
            initialValue: _query.muscle,
            isExpanded: true,
            decoration: const InputDecoration(
              labelText: 'Grupo Muscular',
              filled: true,
              fillColor: Colors.white,
              border: OutlineInputBorder(),
              contentPadding: EdgeInsets.symmetric(horizontal: 10, vertical: 8),
            ),
            items: [
              const DropdownMenuItem(value: '', child: Text('Todos os músculos')),
              ...filters.muscleGroups.map(
                (m) => DropdownMenuItem(value: m.code, child: Text(m.namePT)),
              ),
            ],
            onChanged: (val) => setState(
              () => _query = _query.copyWith(muscle: val ?? '', offset: 0),
            ),
          ),
        ),
        SizedBox(
          width: 180,
          child: DropdownButtonFormField<String>(
            initialValue: _query.equipment,
            isExpanded: true,
            decoration: const InputDecoration(
              labelText: 'Equipamento',
              filled: true,
              fillColor: Colors.white,
              border: OutlineInputBorder(),
              contentPadding: EdgeInsets.symmetric(horizontal: 10, vertical: 8),
            ),
            items: [
              const DropdownMenuItem(value: '', child: Text('Todos equipamentos')),
              ...filters.equipments.map(
                (eq) => DropdownMenuItem(value: eq.code, child: Text(eq.namePT)),
              ),
            ],
            onChanged: (val) => setState(
              () => _query = _query.copyWith(equipment: val ?? '', offset: 0),
            ),
          ),
        ),
        SizedBox(
          width: 180,
          child: DropdownButtonFormField<String>(
            initialValue: _query.category,
            isExpanded: true,
            decoration: const InputDecoration(
              labelText: 'Categoria',
              filled: true,
              fillColor: Colors.white,
              border: OutlineInputBorder(),
              contentPadding: EdgeInsets.symmetric(horizontal: 10, vertical: 8),
            ),
            items: [
              const DropdownMenuItem(value: '', child: Text('Todas categorias')),
              ...filters.categories.map(
                (c) => DropdownMenuItem(value: c.code, child: Text(c.namePT)),
              ),
            ],
            onChanged: (val) => setState(
              () => _query = _query.copyWith(category: val ?? '', offset: 0),
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildExerciseCard(BuildContext context, CatalogExerciseSummary ex) {
    return Card(
      elevation: 0,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: const BorderSide(color: Color(0xFFE4E9EF)),
      ),
      child: InkWell(
        borderRadius: BorderRadius.circular(12),
        onTap: () => context.push('/catalog/${ex.id}'),
        child: Padding(
          padding: const EdgeInsets.all(14),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Expanded(
                    child: Text(
                      ex.namePT,
                      style: const TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                  const Icon(Icons.chevron_right, color: Colors.black45),
                ],
              ),
              if (ex.nameEN.isNotEmpty && ex.nameEN != ex.namePT)
                Padding(
                  padding: const EdgeInsets.only(top: 2),
                  child: Text(
                    ex.nameEN,
                    style: const TextStyle(fontSize: 12, color: Colors.black54),
                  ),
                ),
              const SizedBox(height: 10),
              Wrap(
                spacing: 6,
                runSpacing: 6,
                children: [
                  _tagChip(ex.primaryMusclePT, Colors.blue.shade50, Colors.blue.shade800),
                  _tagChip(ex.equipmentPT, Colors.teal.shade50, Colors.teal.shade800),
                  _tagChip(ex.categoryPT, Colors.orange.shade50, Colors.orange.shade900),
                  _tagChip(
                    ex.isCustom ? 'Customizado' : 'Catálogo Público',
                    ex.isCustom ? Colors.purple.shade50 : Colors.grey.shade200,
                    ex.isCustom ? Colors.purple.shade800 : Colors.black87,
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _tagChip(String label, Color bg, Color fg) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
      decoration: BoxDecoration(
        color: bg,
        borderRadius: BorderRadius.circular(14),
      ),
      child: Text(
        label,
        style: TextStyle(fontSize: 12, color: fg, fontWeight: FontWeight.w600),
      ),
    );
  }
}
