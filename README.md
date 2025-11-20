File: service.test
Type: alloc_space
Time: 2025-11-20 12:11:46 MSK
Showing nodes accounting for 465.79MB, 7.19% of 6475.46MB total
Dropped 25 nodes (cum <= 32.38MB)
      flat  flat%   sum%        cum   cum%
  465.79MB  7.19%  7.19%   465.79MB  7.19%  github.com/Aleksey170999/go-shortener/internal/service.(*memoryURLRepository).GetByUserID
         0     0%  7.19%   465.79MB  7.19%  github.com/Aleksey170999/go-shortener/internal/service.(*URLService).GetUserURLs (inline)
         0     0%  7.19%   466.29MB  7.20%  github.com/Aleksey170999/go-shortener/internal/service.BenchmarkURLService_GetUserURLs
         0     0%  7.19%   465.79MB  7.19%  testing.(*B).launch
         0     0%  7.19%   465.79MB  7.19%  testing.(*B).runN