import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import 'workouts_provider.dart';
import '../auth/auth_provider.dart';

class CreateSessionScreen extends ConsumerStatefulWidget {
  final String planId;
  const CreateSessionScreen({super.key, required this.planId});

  @override
  ConsumerState<CreateSessionScreen> createState() =>
      _CreateSessionScreenState();
}

class _CreateSessionScreenState extends ConsumerState<CreateSessionScreen> {
  final _formKey = GlobalKey<FormState>();

  String _title = '';
  int _dayOfWeek = 1;
  String _notes = '';
  bool _isLoading = false;

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;
    _formKey.currentState!.save();

    setState(() => _isLoading = true);

    try {
      final dio = ref.read(dioClientProvider).dio;
      await dio.post(
        '/workouts/${widget.planId}/sessions',
        data: {
          'workout_plan_id': widget.planId,
          'title': _title,
          'day_of_week': _dayOfWeek,
          'notes': _notes,
        },
      );

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Sessão adicionada com sucesso!'),
            backgroundColor: Colors.green,
          ),
        );
        ref.invalidate(planDetailsProvider(widget.planId));
        context.pop();
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Erro ao salvar: $e'),
            backgroundColor: Colors.red,
          ),
        );
      }
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    const days = [
      'Domingo',
      'Segunda-feira',
      'Terça-feira',
      'Quarta-feira',
      'Quinta-feira',
      'Sexta-feira',
      'Sábado',
    ];

    return Scaffold(
      appBar: AppBar(title: const Text('Nova Sessão de Treino')),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(24.0),
        child: Form(
          key: _formKey,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              TextFormField(
                decoration: const InputDecoration(
                  labelText: 'Título da Sessão (Ex: Treino A - Peito)',
                  border: OutlineInputBorder(),
                ),
                validator: (val) => val == null || val.length < 2
                    ? 'Mínimo de 2 caracteres'
                    : null,
                onSaved: (val) => _title = val ?? '',
              ),
              const SizedBox(height: 16),
              DropdownButtonFormField<int>(
                decoration: const InputDecoration(
                  labelText: 'Dia da Semana',
                  border: OutlineInputBorder(),
                ),
                initialValue: _dayOfWeek,
                items: List.generate(7, (index) {
                  return DropdownMenuItem(
                    value: index,
                    child: Text(days[index]),
                  );
                }),
                onChanged: (val) {
                  if (val != null) setState(() => _dayOfWeek = val);
                },
                onSaved: (val) => _dayOfWeek = val ?? 1,
              ),
              const SizedBox(height: 16),
              TextFormField(
                decoration: const InputDecoration(
                  labelText: 'Anotações (Ex: Foco na execução lenta)',
                  border: OutlineInputBorder(),
                ),
                maxLines: 3,
                onSaved: (val) => _notes = val ?? '',
              ),
              const SizedBox(height: 32),
              SizedBox(
                height: 48,
                child: ElevatedButton(
                  onPressed: _isLoading ? null : _submit,
                  child: _isLoading
                      ? const CircularProgressIndicator(color: Colors.white)
                      : const Text(
                          'Adicionar Sessão',
                          style: TextStyle(fontSize: 16),
                        ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
