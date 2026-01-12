from unittest import TestCase
from tests.Solution import Solution


class TestSolution(TestCase):
    def setUp(self) -> None:
        self.sut = Solution()

    '''Example 1:ц
    Input: m = 3, n = 7
    Output: 28
    '''
    '''Example 2:
    Input: m = 3, n = 2
    Output: 3'''
    def test_login_button(self):
        data = [(3, 7), (3, 2),
                (1, 1), (5, 5), (3, 8), (15, 12), (10, 10)]
        for m, n in data:
            with self.subTest(m=m, n=n):
                self.assertTrue(self.sut.uniquePaths(m, n))
                print(self.sut.uniquePaths(m,n))


    ''' Constraints:
    1 <= m, n <= 100
        0, 1, 2 = m
        99, 100, 101 = n
        У m нет верхнего предела, В Python 3 тип int не имеет максимального ограничения. Тогда просто буду проверять очень большие числа

    answer will be less than or equal to 2 * 10^9
        answer = 2 * 10^9-1, 2 * 10^9, 2 * 10^9 + 1 - Нужно подобрать значения m и n под них (пока не является возможным)
    
    Integers
        - (0.5,7),(8,11.2), (10.0,4) // float
          (!, 1), (@, 1), (#, 1), (%, 1), (2, $),  // string
          (2, !), (2, @), (2, #), (2, %), ($, 2),  // string
          (@, @),(*, *), (а́, 5), (©, 10), (20, >), // string
          (🔥, 🔥), (◀️, ▶️), (α, 4), // string
          ([5,6,4], [7,6,5]), (2 + 3j, 6), ((3, 4), (4, 3)), // complex, list, tuple, 
           ({1, 2, 3, 4, 5}, {1, 2, 3, 4, 5}), (5/3, 3), (4, 6/2),  //set, complex
          ({'name': 'John', 'age': 30, 'city': 'New York'}, {'name': 'John', 'age': 30, 'city': 'New York'}) // dict
        
    Positive_without_0
        + (1,1),(5,5),(50,50),(3,8),(70,12)
        - (-1,5),(0,4),(20,-8),(2,0),(0,0),(-6,-12)
    '''


    def test_password_field(self):
        data = [(-1, 5), (0, 4), (20, -8), (2, 0), (0, 0), (-6, -12)]
        for m, n in data:
            with self.subTest(m=m, n=n):
                with self.assertRaises(ValueError):
                    self.sut.uniquePaths(m, n)


    def test_unique_paths_m_or_n_not_integers(self):
        data = [(0.5,7),(8,11.2), (10.0,4),
                ('!', 1), ('@', 1), ('#', 1), ('%', 1), (2, '$'),
                (2, '!'), (2, '@'), (2, '#'), (2, '%'), ('$', 2),
                ('@', '@'),('*', '*'),
                ('а́', 5), ('©', 10), (20, '>'), ('🔥', '🔥'), ('◀️','▶️'), ('α', 4), (5/3, 3), (4, 6/2),
                ([5, 6, 4], [7, 6, 5]), (2 + 3j, 6), ((3, 4), (4, 3)), ({1, 2, 3, 4, 5}, {1, 2, 3, 4, 5}),
                ({'name': 'John', 'age': 30, 'city': 'New York'}, {'name': 'John', 'age': 30, 'city': 'New York'})]
        for m, n in data:
            with self.subTest(m=m, n=n):
                with self.assertRaises(TypeError):
                    self.sut.uniquePaths(m, n)

    def test_unique_paths_m_and_n_invalid_borders(self):
        data = [(0, 10), (0, 101), (15, 101)]
        for m, n in data:
            with self.subTest(m=m, n=n):
                with self.assertRaises(ValueError):
                    self.sut.uniquePaths(m, n)

    def test_unique_paths_m_and_n_valid_borders(self):
        data = [(1, 10), (2, 15), (1, 100), (5, 99), (1000, 1), (100000, 1), (20000000, 1)]
        for m, n in data:
            with self.subTest(m=m, n=n):
                self.assertTrue(self.sut.uniquePaths(m, n))
                print(self.sut.uniquePaths(m,n))

    def test_unique_paths_big_answer(self):
        data = [(50, 50), (12, 70), (100, 99), (30, 30), (20, 20), ]
        for m, n in data:
            with self.subTest(m=m, n=n):
                with self.assertRaises(ValueError):
                    self.sut.uniquePaths(m, n)

'''# Тест выполняется очень долго из-за долгой работы алгоритма (требований по времени работы алгоритма нет, поэтому не исправлял)
    def test_unique_paths_very_big_m(self):
        data = [(200000000, 2), (2000000000, 2)]
        for m, n in data:
            with self.subTest(m=m, n=n):
                self.assertTrue(self.sut.uniquePaths(m, n))
                with self.assertRaises(TimeoutError):
                    self.sut.uniquePaths(m, n)'''