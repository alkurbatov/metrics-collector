# Инкремент 15
1. Добавьте в свой проект бенчмарки, измеряющие скорость выполнения важнейших компонентов вашей системы.
2. Проведите анализ использования памяти вашим проектом, определите и исправьте неэффективные части кода по следующему алгоритму:
    - Используя профилировщик pprof, сохраните профиль потребления памяти вашим проектом в директорию `profiles` с именем `base.pprof`.
    - Изучите полученный профиль, определите и исправьте неэффективные части вашего кода.
    - Повторите пункт 1 и сохраните новый профиль потребления памяти в директорию `profiles` с именем `result.pprof`.

3. Проверьте результат внесённых изменений командой:
    ```bash
    pprof -top -diff_base=profiles/base.pprof profiles/result.pprof
    ```
    В случае успешной оптимизации вы увидите в выводе командной строки результаты с отрицательными значениями, означающими уменьшение потребления ресурсов.