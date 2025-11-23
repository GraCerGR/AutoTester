class Solution:
    def uniquePaths(self, m: int, n: int) -> int:

        '''ПРОВЕРКИ'''
        if not isinstance(m, int) or not isinstance(n, int):
            raise TypeError("m и n должны быть целыми числами")

        if m < 1 or n < 1:
            raise ValueError("значения m и n должны быть больше или равны 1")

        if n > 100:
            raise ValueError("n должно быть меньше или равно 100")

        #Основной алгоритм
        row = [1] * n
        for i in range(m - 1):
            newRow = [1] * n
            for j in range(n - 2, -1, -1):
                newRow[j] = newRow[j + 1] + row[j]
            row = newRow

        '''ПРОВЕРКА'''
        result = row[0]
        if result > 2 * 10**9:
            raise ValueError("Результат должен быть меньше или равен 2*10^9")


        return result
