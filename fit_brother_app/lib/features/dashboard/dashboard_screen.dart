import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../auth/auth_provider.dart';
import '../students/students_provider.dart';

class DashboardScreen extends ConsumerWidget {
  const DashboardScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final authState = ref.watch(authProvider);
    final isStudent = authState.role == 'STUDENT';

    return Scaffold(
      backgroundColor: const Color(0xFFF5F7FA),
      appBar: AppBar(
        title: const Text('Fit Brother', style: TextStyle(fontWeight: FontWeight.bold)),
        backgroundColor: Colors.white,
        foregroundColor: Colors.black87,
        elevation: 1,
        actions: [
          IconButton(
            icon: const Icon(Icons.logout),
            tooltip: 'Sair',
            onPressed: () async {
              await ref.read(authProvider.notifier).logout();
              if (context.mounted) context.go('/login');
            },
          ),
        ],
      ),
      drawer: _buildDrawer(context, isStudent: isStudent),
      body: isStudent ? _buildStudentHome(context) : _buildTrainerHome(context, ref),
      floatingActionButton: isStudent
          ? null
          : FloatingActionButton.extended(
              onPressed: () => context.push('/students/create'),
              icon: const Icon(Icons.person_add),
              label: const Text('Novo Aluno'),
            ),
    );
  }

  Widget _buildStudentHome(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.all(24),
      child: Center(
        child: Container(
          constraints: const BoxConstraints(maxWidth: 520),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              const Text(
                'Bem-vindo ao Fit Brother',
                style: TextStyle(fontSize: 24, fontWeight: FontWeight.bold),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 12),
              const Text(
                'Acesse seus planos de treino personalizados ou gerencie seus consentimentos LGPD.',
                textAlign: TextAlign.center,
                style: TextStyle(color: Colors.black54),
              ),
              const SizedBox(height: 24),
              ElevatedButton.icon(
                onPressed: () => context.push('/workouts'),
                icon: const Icon(Icons.assignment),
                label: const Text('MEUS PLANOS DE TREINO'),
              ),
              const SizedBox(height: 12),
              OutlinedButton.icon(
                onPressed: () => context.push('/catalog'),
                icon: const Icon(Icons.menu_book),
                label: const Text('BIBLIOTECA DE EXERCÍCIOS'),
              ),
              const SizedBox(height: 12),
              OutlinedButton.icon(
                onPressed: () => context.push('/privacy'),
                icon: const Icon(Icons.privacy_tip_outlined),
                label: const Text('PRIVACIDADE & LGPD'),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildTrainerHome(BuildContext context, WidgetRef ref) {
    final studentsAsync = ref.watch(studentsProvider);

    return Padding(
      padding: const EdgeInsets.all(24),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Wrap(
            alignment: WrapAlignment.spaceBetween,
            crossAxisAlignment: WrapCrossAlignment.center,
            spacing: 12,
            runSpacing: 8,
            children: [
              const Text(
                'Meus Alunos',
                style: TextStyle(fontSize: 22, fontWeight: FontWeight.bold, color: Colors.black87),
              ),
              OutlinedButton.icon(
                onPressed: () => context.push('/catalog'),
                icon: const Icon(Icons.menu_book),
                label: const Text('Biblioteca de Exercícios'),
              ),
            ],
          ),
          const SizedBox(height: 16),
          Expanded(
            child: studentsAsync.when(
              loading: () => const Center(child: CircularProgressIndicator()),
              error: (err, _) => Center(
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    const Icon(Icons.error_outline, color: Colors.red, size: 48),
                    const SizedBox(height: 12),
                    Text(
                      'Erro ao carregar alunos:\n$err',
                      textAlign: TextAlign.center,
                      style: const TextStyle(color: Colors.red),
                    ),
                    const SizedBox(height: 16),
                    ElevatedButton.icon(
                      icon: const Icon(Icons.refresh),
                      label: const Text('Tentar novamente'),
                      onPressed: () => ref.invalidate(studentsProvider),
                    ),
                  ],
                ),
              ),
              data: (students) => students.isEmpty
                  ? const Center(
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Icon(Icons.people_outline, size: 64, color: Colors.grey),
                          SizedBox(height: 12),
                          Text('Nenhum aluno cadastrado ainda.', style: TextStyle(color: Colors.grey)),
                        ],
                      ),
                    )
                  : ListView.separated(
                      itemCount: students.length,
                      separatorBuilder: (context, index) => const SizedBox(height: 8),
                      itemBuilder: (context, index) {
                        final s = students[index];
                        return Card(
                          elevation: 0,
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(12),
                            side: const BorderSide(color: Color(0xFFE8ECF0)),
                          ),
                          child: ListTile(
                            contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                            leading: CircleAvatar(
                              backgroundColor: Colors.blue.shade100,
                              child: Text(
                                s.fullName.isNotEmpty ? s.fullName[0].toUpperCase() : '?',
                                style: const TextStyle(fontWeight: FontWeight.bold, color: Colors.blue),
                              ),
                            ),
                            title: Text(s.fullName, style: const TextStyle(fontWeight: FontWeight.w600)),
                            subtitle: Text(s.email, style: TextStyle(color: Colors.grey.shade600)),
                            onTap: () => context.push('/workouts'),
                            trailing: const Icon(Icons.chevron_right),
                          ),
                        );
                      },
                    ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildDrawer(BuildContext context, {required bool isStudent}) {
    return Drawer(
      child: Column(
        children: [
          Container(
            width: double.infinity,
            padding: const EdgeInsets.fromLTRB(16, 48, 16, 24),
            color: Colors.blue,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const CircleAvatar(
                  radius: 28,
                  backgroundColor: Colors.white,
                  child: Icon(Icons.fitness_center, color: Colors.blue, size: 28),
                ),
                const SizedBox(height: 12),
                Text(
                  isStudent ? 'Área do Aluno' : 'Personal Trainer',
                  style: const TextStyle(color: Colors.white, fontSize: 18, fontWeight: FontWeight.bold),
                ),
              ],
            ),
          ),
          if (!isStudent)
            ListTile(
              leading: const Icon(Icons.people),
              title: const Text('Alunos'),
              selected: true,
              onTap: () => Navigator.pop(context),
            ),
          ListTile(
            leading: const Icon(Icons.assignment),
            title: const Text('Planos de Treino'),
            onTap: () {
              Navigator.pop(context);
              context.push('/workouts');
            },
          ),
          ListTile(
            leading: const Icon(Icons.menu_book),
            title: const Text('Biblioteca de Exercícios'),
            onTap: () {
              Navigator.pop(context);
              context.push('/catalog');
            },
          ),
          if (!isStudent)
            ListTile(
              leading: const Icon(Icons.security),
              title: const Text('Segurança MFA'),
              onTap: () {
                Navigator.pop(context);
                context.push('/settings/security');
              },
            ),
          ListTile(
            leading: const Icon(Icons.privacy_tip_outlined),
            title: const Text('Privacidade & LGPD'),
            onTap: () {
              Navigator.pop(context);
              context.push('/privacy');
            },
          ),
        ],
      ),
    );
  }
}
