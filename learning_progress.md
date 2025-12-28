# Learning Progress

브라우저 구현을 통해 학습하는 모든 개념을 추적합니다. 각 개념은 학습 순서와 관계없이 나열되며, 코드 내 위치를 함께 표시합니다.

**범례**: ✅ 학습 완료 | 🔄 진행 중 | ⬜ 미학습

---

## 1. 네트워킹 기초

### 1.1 URL (Uniform Resource Locator)
- **✅ URL 구조 이해**: 프로토콜(scheme), 호스트(host), 경로(path)로 구성
  - [url.go:12-16](url.go:12) - URL 구조체 정의
- **✅ URL 파싱 (Parsing)**: 문자열을 구조화된 데이터로 변환
  - [url.go:19-53](url.go:19) - NewURL 함수

**주요 개념**:
- `strings.SplitN()`: 문자열을 구분자로 나누기 ([url.go:22](url.go:22))
- `strings.Contains()`: 문자열 포함 여부 확인 ([url.go:36](url.go:36))
- 에러 처리: `error` 타입 반환 ([url.go:24](url.go:24), [28](url.go:28))
- 포인터 반환: `*URL` ([url.go:19](url.go:19))

### 1.2 TCP/IP 통신
- **✅ TCP 연결**: 서버와의 소켓 연결 생성
  - [url.go:60](url.go:60) - `net.Dial("tcp", address)`
- **✅ 데이터 전송**: 바이트 스트림으로 데이터 보내기
  - [url.go:78](url.go:78) - `conn.Write([]byte(request))`
- **✅ 연결 관리**: defer를 사용한 리소스 정리
  - [url.go:65](url.go:65) - `defer conn.Close()`

**주요 개념**:
- `net.Dial()`: 네트워크 연결 생성
- `defer`: 함수 종료 시 실행할 코드 예약
- 타입 변환: `[]byte()` string을 byte slice로 변환

### 1.3 HTTP 프로토콜
- **✅ HTTP 요청 메시지 구성**: GET 메서드로 리소스 요청
  - [url.go:69-74](url.go:69) - HTTP/1.0 요청 포맷
  - [url_llm.go:124-150](url_llm.go:124) - HTTP/1.1 요청 포맷 (개선)
- **✅ HTTP 응답 파싱**: 상태 라인, 헤더, 바디 분리
  - [url.go:86-113](url.go:86) - 응답 읽기 로직
- **✅ HTTP 헤더 관리**: 확장 가능한 헤더 구조
  - [url_llm.go:126-130](url_llm.go:126) - map으로 헤더 관리

**주요 개념**:
- HTTP 요청 구조: 메서드, 경로, 프로토콜 버전, 헤더
- `\r\n`: HTTP 프로토콜의 줄바꿈 (CRLF)
- `bufio.NewReader()`: 버퍼링된 읽기 ([url.go:86](url.go:86))
- `io.ReadAll()`: 모든 데이터 읽기 ([url.go:108](url.go:108))
- `map[string]string`: 키-값 쌍으로 헤더 관리
- `strings.Builder`: 효율적인 문자열 조합 ([url_llm.go:136-142](url_llm.go:136))

### 1.4 고급 네트워킹
- **✅ HTTPS/TLS**: 암호화된 통신 ([fetcher.go:206,208](fetcher.go:206))
- **🔄 HTTP/1.1 Keep-Alive**: 연결 재사용으로 성능 향상
  - ✅ ConnectionPool 패턴 구현 ([fetcher.go:37-104](fetcher.go:37))
  - ✅ Content-Length 기반 body 읽기 ([fetcher.go:306-323](fetcher.go:306))
  - ⬜ Transfer-Encoding: chunked (다음 작업)
- ⬜ **리다이렉트 처리**: 3xx 응답 코드 처리
- ⬜ **쿠키 관리**: Cookie 헤더 처리
- ⬜ **캐싱**: 응답 캐시 및 재사용

**주요 개념 (Keep-Alive)**:
- TCP 3-way handshake 비용 절감
- 연결 풀(Pool) 관리: Check-out/Check-in 패턴
- `sync.Mutex`: 동시성 제어 ([fetcher.go:39](fetcher.go:39))
- `io.ReadFull()`: 정확한 바이트 수 읽기 ([fetcher.go:318](fetcher.go:318))
- LIFO 전략: 마지막 사용 연결 우선 재사용

---

## 2. HTML 파싱

### 2.1 HTML 구조
- **✅ HTML 태그 제거**: `<>` 태그를 인식하고 제거하여 텍스트만 추출
  - [url.go:116-128](url.go:116) - show 함수 구현
  - [show_llm.go:119-139](show_llm.go:119) - 학습용 버전
- ⬜ **HTML 문법**: 태그, 속성, 중첩 구조 완전 파싱
- ⬜ **DOM (Document Object Model)**: 트리 구조로 표현

**주요 개념**:
- 상태 플래그: `inTag` 변수로 태그 안/밖 추적
- `<` 감지: 태그 시작
- `>` 감지: 태그 종료
- 조건부 출력: 태그 밖의 문자만 출력

### 2.2 렉싱/토큰화 (Lexing/Tokenization)
- ⬜ **토큰 정의**: 시작 태그, 종료 태그, 텍스트, 주석
- ⬜ **렉서 구현**: HTML 문자열을 토큰 스트림으로 변환

### 2.3 파싱 (Parsing)
- ⬜ **파서 구현**: 토큰 스트림을 DOM 트리로 변환
- ⬜ **에러 처리**: 잘못된 HTML 복구
- ⬜ **특수 태그 처리**: `<script>`, `<style>`, `<pre>` 등

---

## 3. CSS 파싱 및 스타일링

### 3.1 CSS 구조
- ⬜ **CSS 문법**: 선택자, 속성, 값
- ⬜ **선택자 종류**: 태그, 클래스, ID, 속성, 가상 선택자

### 3.2 CSS 파싱
- ⬜ **CSS 렉싱**: CSS 문자열을 토큰으로 변환
- ⬜ **규칙 파싱**: 선택자와 선언 블록 분리

### 3.3 스타일 계산
- ⬜ **선택자 매칭**: DOM 노드에 CSS 규칙 적용
- ⬜ **명시도 (Specificity)**: 우선순위 계산
- ⬜ **상속 (Inheritance)**: 부모 스타일 상속
- ⬜ **계산된 값 (Computed Values)**: 최종 스타일 값 계산

---

## 4. 레이아웃 (Layout)

### 4.1 박스 모델 (Box Model)
- ⬜ **박스 차원**: content, padding, border, margin
- ⬜ **박스 타입**: block, inline, inline-block

### 4.2 레이아웃 알고리즘
- ⬜ **플로우 레이아웃**: 블록 및 인라인 요소 배치
- ⬜ **너비/높이 계산**: 박스 크기 결정
- ⬜ **위치 계산**: 좌표 결정

### 4.3 고급 레이아웃 (미학습)
- ⬜ **Flexbox**: 유연한 박스 레이아웃
- ⬜ **Grid**: 그리드 레이아웃
- ⬜ **Positioning**: absolute, relative, fixed

---

## 5. 렌더링 (Rendering)

### 5.1 렌더 트리
- ⬜ **렌더 트리 구축**: DOM + CSSOM → Render Tree
- ⬜ **표시 요소 필터링**: `display: none` 제외

### 5.2 페인팅 (Painting)
- ⬜ **그래픽 출력**: 화면에 픽셀 그리기
- ⬜ **텍스트 렌더링**: 글꼴 및 문자 출력
- ⬜ **배경 및 테두리**: 색상 및 이미지 그리기

### 5.3 디스플레이
- ⬜ **화면 출력**: 터미널 또는 GUI에 표시
- ⬜ **스크롤**: 뷰포트 및 스크롤 처리

---

## 6. Go 언어 기초 개념

### 6.1 구조체 및 메서드
- **✅ 구조체 정의**: `type URL struct` ([url.go:12-16](url.go:12))
- **✅ 메서드**: 구조체에 함수 연결 ([url.go:56](url.go:56) - `func (u *URL) Request()`)
- **✅ 포인터 리시버**: `*URL`로 메서드 정의

### 6.2 에러 처리
- **✅ 에러 반환**: 함수가 `error` 타입 반환 ([url.go:19](url.go:19), [56](url.go:56))
- **✅ 에러 생성**: `fmt.Errorf()` 사용 ([url.go:24](url.go:24), [28](url.go:28))
- **✅ 에러 체크**: `if err != nil` 패턴 ([url.go:61-63](url.go:61), [79-81](url.go:79), [90-92](url.go:90))

### 6.3 패키지 및 임포트
- **✅ 표준 라이브러리 사용**: `fmt`, `net`, `strings`, `bufio`, `io`, `os` ([url.go:3-9](url.go:3))
  - `os` 패키지: 운영체제 기능, 커맨드 라인 인자 접근
- ⬜ **서드파티 패키지**: 외부 라이브러리 사용
- ⬜ **모듈 관리**: `go.mod`, `go.sum`

### 6.4 문자열과 타입
- **✅ rune 타입**: 유니코드 코드포인트를 나타내는 int32 별칭
  - `for _, c := range body`: 문자열을 rune으로 순회 ([url.go:119](url.go:119))
  - `string(c)`: rune을 string으로 변환 ([url.go:125](url.go:125))
- **✅ 문자 vs 문자열 리터럴**:
  - `'<'`: 작은따옴표 = 문자 (rune 타입)
  - `"<"`: 큰따옴표 = 문자열 (string 타입)
- **✅ range를 사용한 순회**:
  - `for _, c := range body`: 인덱스 무시, 각 rune 순회

### 6.5 커맨드 라인 인자
- **✅ os.Args**: 커맨드 라인 인자 배열
  - `os.Args[0]`: 프로그램 이름
  - `os.Args[1]`: 첫 번째 인자
  - `len(os.Args)`: 인자 개수 확인 ([url.go:135](url.go:135))
- **✅ 인자 유효성 검사**: 인자 부족 시 사용법 출력

**파이썬과 비교**:
- 파이썬: `sys.argv`, `if __name__ == "__main__":`
- Go: `os.Args`, main 함수가 항상 진입점

### 6.6 함수 추상화
- **✅ load 함수**: 여러 단계를 하나로 통합 ([url.go:130-143](url.go:130))
  - URL 파싱 + HTTP 요청 + 출력을 한 함수로

### 6.7 동시성 (Concurrency)
- **✅ Mutex (뮤텍스)**: 공유 자원 보호
  - `sync.Mutex`: 상호 배제 잠금 ([fetcher.go:39](fetcher.go:39))
  - `Lock()` / `Unlock()`: 크리티컬 섹션 보호
  - `defer Unlock()`: 패닉 시에도 잠금 해제 보장
  - 경쟁 조건(Race Condition) 방지
- ⬜ **고루틴 (Goroutines)**: 경량 스레드
- ⬜ **채널 (Channels)**: 고루틴 간 통신
- ⬜ **WaitGroup**: 고루틴 완료 대기

### 6.8 테스팅
- **✅ 단위 테스트 (Unit Testing)**: 개별 함수 테스트
  - [parser_test.go:5-63](parser_test.go:5) - parseHTML 테스트
  - [url_test.go:69-224](url_test.go:69) - NewURL 테스트
  - [url_test.go:230-335](url_test.go:230) - parsePort 테스트
  - [url_test.go:341-429](url_test.go:341) - parseHostPath 테스트
  - [fetcher_test.go:238-347](fetcher_test.go:238) - HTTPFetcher 테스트
  - [fetcher_test.go:435-507](fetcher_test.go:435) - ConnectionPool 테스트
- **✅ Mock 객체**: 테스트용 가짜 구현
  - `mockConn`: net.Conn 인터페이스 구현 ([fetcher_test.go:351-393](fetcher_test.go:351))
  - `httptest.NewServer`: Mock HTTP 서버 ([fetcher_test.go:240](fetcher_test.go:240))
- **✅ 테스트 함수 작성**: `func TestXxx(t *testing.T)`
  - `t.Errorf()`: 테스트 실패 보고
  - `t.Fatalf()`: 치명적 에러로 즉시 종료
- **✅ 테스트 실행**: `go test -v`

**주요 개념**:
- 테스트 파일: `_test.go` 접미사
- 테스트 함수 네이밍: `Test` + 함수명 + `_` + 시나리오
- AAA 패턴: Arrange(준비), Act(실행), Assert(검증)
- 에러 케이스 테스트: 잘못된 입력 검증
- Mock 패턴: 외부 의존성 제거

### 6.9 인터페이스와 다형성
- **✅ 인터페이스 정의**: `type Fetcher interface` ([browser.go:56-58](browser.go:56))
  - 메서드 시그니처만 선언: `Fetch(u *URL) (string, error)`
- **✅ 인터페이스 구현**: 구조체에 메서드 추가
  - `FileFetcher`: file:// 프로토콜 처리 ([browser.go:159-173](browser.go:159))
  - `DataFetcher`: data:// 프로토콜 처리 ([browser.go:175-203](browser.go:175))
  - `HTTPFetcher`: http://, https:// 프로토콜 처리 ([browser.go:205-282](browser.go:205))
- **✅ Registry 패턴**: map으로 구현체 관리
  - `fetcherRegistry`: scheme별 Fetcher 등록 ([browser.go:143-148](browser.go:143))
  - 동적 디스패치: scheme에 따라 적절한 Fetcher 선택
- **✅ 개방-폐쇄 원칙 (OCP)**: 새 프로토콜 추가 시 기존 코드 수정 불필요

**주요 개념**:
- 암묵적 인터페이스 구현 (implicit interface)
- 덕 타이핑 (duck typing): "Fetch 메서드가 있으면 Fetcher"
- 다형성: 모든 Fetcher를 동일하게 취급

### 6.10 함수 리팩토링과 순수 함수
- **✅ 순수 함수 설계**: 부작용 없는 함수
  - `parsePort()`: scheme/host → cleanHost/port/error ([browser.go:142-178](browser.go:142))
  - `parseHostPath()`: scheme/rest → host/path ([browser.go:180-201](browser.go:180))
  - `parseHTML()`: HTML → 텍스트 ([browser.go:284-303](browser.go:284))
- **✅ 함수 분해**: 큰 함수를 작은 함수로 분리
  - NewURL: 60줄 → 30줄 (50% 감소)
  - 중첩 if문 제거: 3단계 → 1단계
- **✅ 단일 책임 원칙 (SRP)**: 함수가 한 가지만 수행

**주요 개념**:
- 순수 함수: 같은 입력 → 같은 출력, 부작용 없음
- 함수 시그니처 설계: 명확한 입력/출력
- 명명된 반환값: `(cleanHost string, port int, err error)`
- 테스트 가능성: 순수 함수는 테스트하기 쉬움

### 6.11 고급 기능
- ⬜ **제네릭**: 타입 파라미터
- ⬜ **리플렉션**: 런타임 타입 검사

---

## 학습 로드맵 (밑바닥부터 시작하는 웹 브라우저)

### PART 1: 텍스트 표시

#### CHAPTER 1: 웹페이지 다운로드 ✅
**시작**: 2024-12-24 | **완료**: 2024-12-24

- [x] 1.1 서버에 연결하기 - 2024-12-24
  - TCP 연결, `net.Dial()` ([url.go:60](url.go:60))
- [x] 1.2 정보 요청하기 - 2024-12-24
  - HTTP GET 요청 구성 ([url.go:69-74](url.go:69))
- [x] 1.3 서버의 응답 - 2024-12-24
  - 상태 라인, 헤더, 바디 파싱 ([url.go:86-113](url.go:86))
- [x] 1.4 파이썬을 통한 텔넷 - 2024-12-24
  - URL 파싱, NewURL 함수 ([url.go:19-53](url.go:19))
- [x] 1.5 요청과 응답 - 2024-12-24
  - Request 메서드 통합 ([url.go:56-114](url.go:56))
- [x] 1.6 HTML 표시하기 - 2024-12-24
  - show 함수로 태그 제거 ([url.go:116-128](url.go:116))
  - load 함수로 통합 ([url.go:130-143](url.go:130))
  - 커맨드 라인 인자 처리 ([url.go:135-138](url.go:135))
- [x] 1.7 암호화된 연결 - 2024-12-24
  - HTTPS/TLS 지원 ([show_llm.go:66-78](show_llm.go:66))
  - `crypto/tls` 패키지 사용
  - `tls.Dial()` vs `net.Dial()` 조건부 처리
  - 포트 443 (HTTPS) vs 80 (HTTP)
  - defer 에러 처리 ([show_llm.go:86-92](show_llm.go:86))
- [ ] 1.8 요약
- [x] 1.9 연습 문제 - 2025-12-25
  - **HTTP/1.1 지원**: HTTP/1.0에서 HTTP/1.1로 업그레이드
  - **Connection 헤더**: `Connection: close` 추가
  - **User-Agent 헤더**: `User-Agent: GoWebBrowser/1.0` 추가
  - **확장 가능한 헤더 구조**: map과 strings.Builder 사용 ([url_llm.go:124-150](url_llm.go:124))

#### CHAPTER 2: 화면에 그리기 ⬜
- [ ] 2.1 창 만들기
- [ ] 2.2 창에 그리기
- [ ] 2.3 텍스트 배치하기
- [ ] 2.4 텍스트 스크롤하기
- [ ] 2.5 더 빠른 렌더링
- [ ] 2.6 요약
- [ ] 2.7 연습 문제

#### CHAPTER 3: 텍스트 포맷팅하기 ⬜
- [ ] 3.1 폰트(서체)란?
- [ ] 3.2 텍스트 측정하기
- [ ] 3.3 한 단어씩 처리하기
- [ ] 3.4 텍스트에 스타일 주기
- [ ] 3.5 레이아웃 객체
- [ ] 3.6 다양한 크기의 텍스트
- [ ] 3.7 폰트 캐싱
- [ ] 3.8 요약
- [ ] 3.9 연습 문제

---

### PART 2: 문서 표시

#### CHAPTER 4: 문서 트리 구축하기 ⬜
- [ ] 4.1 노드 트리
- [ ] 4.2 트리 구축하기
- [ ] 4.3 파서 디버깅하기
- [ ] 4.4 셀프 클로징 태그
- [ ] 4.5 노드 트리 사용하기
- [ ] 4.6 페이지 오류 다루기
- [ ] 4.7 요약
- [ ] 4.8 연습 문제

#### CHAPTER 5: 페이지 레이아웃 ⬜
- [ ] 5.1 레이아웃 트리
- [ ] 5.2 블록 레이아웃
- [ ] 5.3 크기와 위치
- [ ] 5.4 재귀 페인팅
- [ ] 5.5 배경 그리기
- [ ] 5.6 요약
- [ ] 5.7 연습 문제

#### CHAPTER 6: 개발자 스타일 적용하기 ⬜
- [ ] 6.1 함수를 사용한 파싱
- [ ] 6.2 style 어트리뷰트
- [ ] 6.3 셀렉터
- [ ] 6.4 스타일시트 적용하기
- [ ] 6.5 캐스케이딩
- [ ] 6.6 상속된 스타일
- [ ] 6.7 폰트 프로퍼티
- [ ] 6.8 요약
- [ ] 6.9 연습 문제

#### CHAPTER 7: 버튼과 링크 처리하기 ⬜
- [ ] 7.1 링크는 어디에 있는가?
- [ ] 7.2 라인 레이아웃
- [ ] 7.3 클릭 처리
- [ ] 7.4 탭 브라우징
- [ ] 7.5 브라우저 크롬
- [ ] 7.6 히스토리 탐색
- [ ] 7.7 URL 입력하기
- [ ] 7.8 요약
- [ ] 7.9 연습 문제

---

### PART 3: 애플리케이션 실행

#### CHAPTER 8: 서버로 정보 보내기 ⬜
- [ ] 8.1 폼의 동작 방식
- [ ] 8.2 위젯을 렌더링하기
- [ ] 8.3 위젯과 상호작용하기
- [ ] 8.4 폼을 제출하기
- [ ] 8.5 웹 앱의 동작
- [ ] 8.6 POST 요청 수신하기
- [ ] 8.7 웹페이지 생성하기
- [ ] 8.8 요약
- [ ] 8.9 연습 문제

#### CHAPTER 9: 대화형 스크립트 실행하기 ⬜
- [ ] 9.1 DukPy 설치하기
- [ ] 9.2 자바스크립트 코드 실행하기
- [ ] 9.3 함수 익스포트하기
- [ ] 9.4 크래시 처리하기
- [ ] 9.5 핸들 반환하기
- [ ] 9.6 핸들 래핑
- [ ] 9.7 이벤트 처리
- [ ] 9.8 DOM 수정하기
- [ ] 9.9 이벤트 기본값
- [ ] 9.10 요약
- [ ] 9.11 연습 문제

#### CHAPTER 10: 사용자 데이터 보호하기 ⬜
- [ ] 10.1 쿠키
- [ ] 10.2 로그인 시스템
- [ ] 10.3 쿠키 구현하기
- [ ] 10.4 교차 사이트 요청
- [ ] 10.5 동일 출처 정책
- [ ] 10.6 교차 사이트 요청 위조
- [ ] 10.7 SameSite 쿠키
- [ ] 10.8 교차 사이트 스크립팅
- [ ] 10.9 콘텐츠 보안 정책
- [ ] 10.10 요약
- [ ] 10.11 연습 문제

---

### PART 4: 모던 브라우저 기능

#### CHAPTER 11: 시각적 효과 ⬜
#### CHAPTER 12: 태스크와 스레드 스케줄링 ⬜
#### CHAPTER 13: 애니메이션과 컴포지팅 ⬜
#### CHAPTER 14: 콘텐츠 접근성 향상 ⬜
#### CHAPTER 15: 삽입된 콘텐츠 지원 ⬜
#### CHAPTER 16: 이전 결과 재사용 ⬜
#### CHAPTER 17: 이 책에서 다루지 않은 내용 ⬜

---

### 진행 상황 요약

**완료**:
- CHAPTER 1 전체 (1.1 ~ 1.7) - 2024-12-24 ✅
- CHAPTER 1 연습 문제 - HTTP/1.1 및 헤더 개선 - 2025-12-25 ✅
- HTTP Keep-Alive 구현 (부분 완료) - 2025-12-28 🔄
  - ✅ ConnectionPool 패턴
  - ✅ Content-Length 기반 읽기
  - ⬜ Transfer-Encoding: chunked (다음 작업)

**다음**: Transfer-Encoding: chunked 구현
**전체 진행률**: 7/152 섹션 완료 (~5%), HTTP/1.1 고급 기능 진행 중

---

## 학습 노트

### 2024-12-24: CHAPTER 1 완료 - 웹페이지 다운로드

#### Phase 1: 기본 HTTP 클라이언트 (1.1 ~ 1.6)
- Go 언어로 웹 브라우저 구현 시작
- 기본 HTTP 클라이언트 완성 ([url.go:1](url.go:1))
- URL 파싱, TCP 연결, HTTP 요청/응답 처리 학습 완료
- **HTML 태그 제거 구현** ([url.go:116-128](url.go:116))
  - `show` 함수: 상태 플래그로 `<>` 태그 감지 및 제거
  - `load` 함수: URL 파싱 → 요청 → 출력 통합
  - 커맨드 라인 인자 처리: `os.Args`로 동적 URL 입력

#### Phase 2: HTTPS/TLS 지원 (1.7)
- **암호화된 연결 구현** ([show_llm.go:66-78](show_llm.go:66))
  - `crypto/tls` 패키지 사용
  - Scheme에 따른 조건부 연결: HTTP(80) vs HTTPS(443)
  - `tls.Dial()`: TLS 핸드셰이크 자동 수행
  - 인증서 검증 자동화
- **에러 처리 개선** ([show_llm.go:86-92](show_llm.go:86))
  - defer + 익명 함수로 `Close()` 에러 처리
  - 변수 shadowing 방지 (`closeErr` 사용)

#### Go 언어 개념 학습
- **문자열 처리**:
  - rune 타입과 문자열 순회 (`for range`)
  - 문자 리터럴(`'<'`) vs 문자열 리터럴(`"<"`)
- **네트워크 프로그래밍**:
  - `net` 패키지: TCP 연결
  - `crypto/tls` 패키지: TLS 암호화
  - `net.Conn` 인터페이스: 다형성
- **시스템 프로그래밍**:
  - `os` 패키지와 커맨드 라인 인자
  - defer를 사용한 리소스 관리
  - 에러 처리 모범 사례
- **설계 패턴**:
  - 함수 추상화와 코드 재사용
  - 조건부 로직으로 다중 프로토콜 지원

#### 학습 환경 설정
- **LLM 파일 네이밍 규칙**: `_llm` postfix로 충돌 방지
- **Wrapup 규칙**: 완료 시 `learning_progress.md` 자동 업데이트

---

### 2025-12-25: CHAPTER 1 연습 문제 - HTTP/1.1 및 헤더 개선

#### HTTP/1.1 지원 구현
- **HTTP/1.0 → HTTP/1.1 업그레이드** ([url_llm.go:133](url_llm.go:133))
  - Request Line: `GET /path HTTP/1.1`
  - HTTP/1.1 기능 기반 마련
- **필수 헤더 추가**:
  - `Host`: 서버 호스트 이름 (HTTP/1.1 필수)
  - `Connection: close`: 연결 종료 명시
  - `User-Agent: GoWebBrowser/1.0`: 브라우저 식별

#### 확장 가능한 헤더 구조 설계
- **map[string]string으로 헤더 관리** ([url_llm.go:126-130](url_llm.go:126))
  - 키-값 쌍으로 헤더 저장
  - 나중에 헤더 추가/제거가 쉬움
  - 동적 헤더 관리 가능
- **strings.Builder로 효율적 문자열 조합** ([url_llm.go:136-142](url_llm.go:136))
  - 문자열 반복 연결 시 메모리 효율적
  - `+` 연산자보다 성능 우수
  - `WriteString()` 메서드로 순차 조합

#### Go 언어 개념 학습
- **컬렉션 타입**:
  - `map[string]string`: 키-값 저장소
  - `for key, value := range map`: 맵 순회
- **문자열 빌더**:
  - `strings.Builder`: 가변 문자열 버퍼
  - `WriteString()`: 문자열 추가
  - `String()`: 최종 문자열 반환
- **코드 구조화**:
  - 데이터 구조와 로직 분리
  - 확장 가능한 설계 패턴

#### 다음 개선 아이디어 (보류)
- URL 구조체에 Headers 필드 추가
- Request 메서드에 옵션 파라미터
- 헤더 추가 메서드 (AddHeader)
- 빌더 패턴 (WithHeader)

---

### 2025-12-27: 인터페이스 리팩토링 및 테스트 작성

#### Fetcher 인터페이스 도입
- **인터페이스 기반 리팩토링** ([browser.go:56-58](browser.go:56))
  - `Fetcher` 인터페이스: 모든 프로토콜 통합
  - `FileFetcher`, `DataFetcher`, `HTTPFetcher` 구현
  - Registry 패턴으로 scheme별 Fetcher 관리 ([browser.go:143-148](browser.go:143))
- **개방-폐쇄 원칙 적용**:
  - 새 프로토콜 추가 시 `fetcherRegistry`에만 등록
  - 기존 코드 수정 불필요
  - 테스트 시 Mock Fetcher 쉽게 작성 가능

#### 종합 테스트 작성
**총 25개 테스트 작성 (모두 통과 ✅)**

1. **NewURL 테스트** ([browser_test.go:69-224](browser_test.go:69))
   - HTTP/HTTPS 기본/커스텀 포트
   - File URL (Windows 경로)
   - Data URL
   - 에러 케이스 (잘못된 scheme, 누락된 scheme)

2. **parsePort 테스트** ([browser_test.go:230-335](browser_test.go:230))
   - scheme별 기본 포트 (HTTP:80, HTTPS:443)
   - 커스텀 포트 파싱
   - file 스킴 (포트 없음)
   - 잘못된 포트 번호 에러 처리

3. **parseHostPath 테스트** ([browser_test.go:341-429](browser_test.go:341))
   - HTTP/HTTPS host/path 분리
   - 경로 없는 URL (기본 "/")
   - 포트 포함 URL
   - file 스킴 절대/상대 경로

#### 함수 리팩토링
- **parsePort 함수 분리** ([browser.go:142-178](browser.go:142))
  - 포트 파싱 로직을 순수 함수로 추출
  - file 스킴 버그 수정 (포트가 80이 아닌 0으로 설정)
  - 명명된 반환값으로 가독성 향상

- **parseHostPath 함수 분리** ([browser.go:180-201](browser.go:180))
  - host/path 파싱 로직을 순수 함수로 추출
  - scheme별 분기 로직 캡슐화
  - NewURL 함수 60줄 → 30줄 (50% 감소)

- **NewURL 간소화**:
  - Before: 3단계 중첩 if문, 약 60줄
  - After: 명확한 5단계 처리, 약 30줄
  ```go
  1. data 스킴 특별 처리
  2. scheme 파싱 및 검증
  3. host/path 분리 ← parseHostPath()
  4. 포트 파싱 ← parsePort()
  5. URL 생성 및 반환
  ```

#### Go 언어 개념 학습
- **인터페이스와 다형성**:
  - 암묵적 인터페이스 구현 (no `implements` keyword)
  - 덕 타이핑: 메서드 시그니처만 일치하면 구현
  - 인터페이스 변수에 구현체 할당

- **테스트 작성**:
  - `testing` 패키지 사용
  - `t.Errorf()` vs `t.Fatalf()` 차이
  - AAA 패턴 (Arrange-Act-Assert)
  - 테스트 네이밍 컨벤션

- **순수 함수 설계**:
  - 부작용 없는 함수
  - 테스트 가능성 극대화
  - 명명된 반환값으로 명확성 향상

#### 설계 원칙 적용
- **단일 책임 원칙 (SRP)**: 각 함수가 한 가지만 수행
- **개방-폐쇄 원칙 (OCP)**: 확장에 열려있고 수정에 닫혀있음
- **의존성 역전 원칙 (DIP)**: 구체적 구현이 아닌 인터페이스에 의존

#### 버그 수정
- **file 스킴 포트 버그**: 테스트를 통해 발견
  - 문제: file URL의 포트가 0이 아닌 80으로 설정됨
  - 원인: 포트 파싱 로직에서 file 스킴 제외하지 않음
  - 해결: parsePort에서 file 스킴 early return

#### Fetcher 통합 테스트 작성
**총 15개 Fetcher 테스트 추가 (모두 통과 ✅)**

1. **FileFetcher 테스트** ([fetcher_test.go:14-90](fetcher_test.go:14))
   - testdata/simple.html 읽기 테스트
   - testdata/empty.html 빈 파일 테스트
   - testdata/entities.html HTML 엔티티 테스트
   - 존재하지 않는 파일 에러 처리

2. **DataFetcher 테스트** ([fetcher_test.go:112-226](fetcher_test.go:112))
   - 일반 텍스트 data URL
   - base64 인코딩된 data URL
   - URL 인코딩된 data URL (%20 등)
   - 복잡한 HTML data URL
   - 에러 케이스 (쉼표 없음, 잘못된 base64)

3. **HTTPFetcher 테스트** ([fetcher_test.go:238-347](fetcher_test.go:238))
   - httptest.NewServer를 사용한 Mock HTTP 서버
   - 성공적인 HTTP 요청 테스트
   - 경로가 있는 HTTP 요청
   - 빈 응답 처리
   - HTTPS URL 파싱 검증 (실제 요청은 skip)
   - 존재하지 않는 호스트 에러 처리

#### 코드베이스 구조 개선
**Feature-based 파일 분리 (Option 1 적용)**

- **모놀리식 구조 문제**:
  - browser.go가 400줄 이상으로 비대화
  - 모든 기능이 한 파일에 혼재
  - 테스트 파일도 700줄 이상 (40개 테스트)

- **파일 분리 완료**:
  ```
  browser.go (400줄)
    ↓
  ├── url.go         (143줄) - URL 파싱 로직
  ├── fetcher.go     (191줄) - Fetcher 인터페이스 및 구현
  ├── parser.go       (35줄) - HTML 파싱 로직
  └── browser.go      (43줄) - main() 및 load() 함수만
  ```

- **테스트 파일 분리 완료**:
  ```
  browser_test.go (776줄)
    ↓
  ├── parser_test.go  (68줄)  - parseHTML 테스트 (5개)
  ├── url_test.go    (397줄)  - URL 관련 테스트 (20개)
  └── fetcher_test.go (343줄) - Fetcher 테스트 (15개)
  ```

- **Symlink 구조 확립**:
  - `llm/` 디렉토리에서 루트 테스트 파일을 symlink로 참조
  - `llm/testdata/` → `../testdata` (기존)
  - `llm/parser_test.go` → `../parser_test.go` (신규)
  - `llm/url_test.go` → `../url_test.go` (신규)
  - `llm/fetcher_test.go` → `../fetcher_test.go` (신규)

#### Go 패키지 구조 학습
- **같은 패키지, 여러 파일**:
  - 모든 파일이 `package main`
  - 파일 간 import 불필요
  - 빌드 시 모든 .go 파일 자동 포함
  - 네임스페이스 충돌 없음

- **파일 분리 기준**:
  - 기능별 분리 (Feature-based)
  - 각 파일이 명확한 책임 영역
  - 테스트 파일도 동일한 기준으로 분리

#### 설계 결정
- **Option 1 선택**: Feature-based 파일 분리
  - 장점: 관련 기능이 한 곳에 모임
  - 장점: 확장성 좋음 (새 기능 추가 시 새 파일 생성)
  - 장점: 학습 프로젝트에 적합 (개념별 분리)

- **대안 (보류)**:
  - Option 2: Fetcher별 파일 분리 (너무 세분화)
  - Option 3: 패키지 분리 (현재 규모에 과함)

#### 최종 테스트 현황
- **총 40개 테스트** (39 pass, 1 skip)
  - parseHTML: 5개
  - NewURL: 8개
  - parsePort: 6개
  - parseHostPath: 6개
  - FileFetcher: 4개
  - DataFetcher: 6개
  - HTTPFetcher: 5개

---

### 2025-12-28: HTTP Keep-Alive 구현 및 리팩토링

#### Phase 1: Keep-Alive 기초 이해
- **문제 인식**: 매 요청마다 TCP 3-way handshake 반복
  - 연결 생성 비용: SYN → SYN-ACK → ACK (3번 왕복)
  - 연결 종료 비용: FIN → ACK → FIN → ACK (4번 왕복)
  - 지연 시간(latency) 증가
- **Keep-Alive 개념 학습**:
  - HTTP/1.1 기본 동작: 연결 재사용
  - `Connection: close` 제거하면 keep-alive 활성화
  - `Content-Length` 필요: 응답 끝 판단

#### Phase 2: Content-Length 기반 Body 읽기
- **문제**: `io.ReadAll()`은 EOF까지 읽음 → 연결 재사용 불가
- **해결**: `Content-Length` 헤더 파싱 후 정확한 바이트 수만 읽기
  - `strconv.Atoi()`: 문자열 → 정수 변환 ([fetcher.go:311](fetcher.go:311))
  - `io.ReadFull()`: 정확히 N바이트 읽기 ([fetcher.go:318](fetcher.go:318))
  - 연결이 닫히지 않아 재사용 가능
- **헤더 파싱 개선** ([fetcher.go:271-294](fetcher.go:271)):
  - map으로 모든 헤더 저장
  - `Content-Length` 존재 여부 확인

#### Phase 3: ConnectionPool 구현
- **초기 설계**: map[string]net.Conn (연결 1개만 저장)
- **문제 발견**: 동시 요청 시 같은 연결 재사용 → 데이터 섞임
- **개선**: Array-based Pool ([fetcher.go:37-104](fetcher.go:37))
  ```go
  type ConnectionPool struct {
      connections map[string][]net.Conn  // host:port → 여러 연결
      mu          sync.Mutex              // 동시성 제어
      maxPerHost  int                     // 최대 6개 (RFC 2616)
  }
  ```
- **Check-out/Check-in 패턴**:
  - `Get()`: Pool에서 연결 꺼내기 (LIFO)
  - `Put()`: Pool에 연결 반납하기
  - 초과 연결은 즉시 닫기 (메모리 누수 방지)

#### Phase 4: 리팩토링 및 Best Practices
- **Logging 시스템 도입** ([fetcher.go:42-53](fetcher.go:42)):
  - `log.Logger` 사용
  - `DEBUG` 환경 변수로 제어
  - 프로덕션: `io.Discard`로 silent
  - 개발: `DEBUG=1`로 상세 로그
- **Godoc 추가**:
  - 패키지 레벨 문서
  - 모든 public 타입/함수에 문서화
  - 반환값 설명: named returns 사용
- **parseResponse 개선** ([fetcher.go:296-376](fetcher.go:296)):
  - 반환값 추가: `(body, headers, error)`
  - 테스트 가능성 향상

#### Phase 5: ConnectionPool 테스트 작성
**총 4개 테스트 추가 (모두 통과 ✅)**

1. **TestConnectionPool_GetPut** ([fetcher_test.go:435-456](fetcher_test.go:435))
   - 기본 Get/Put 동작 검증
   - 빈 Pool에서 Get → nil 반환
   - Put 후 Get → 연결 반환

2. **TestConnectionPool_MaxPerHost** ([fetcher_test.go:458-472](fetcher_test.go:458))
   - 최대 6개 연결 제한 검증
   - 7번째 연결은 자동으로 닫힘

3. **TestConnectionPool_MultipleHosts** ([fetcher_test.go:474-489](fetcher_test.go:474))
   - 여러 호스트 독립적 관리
   - host1의 연결이 host2에 영향 없음

4. **TestConnectionPool_Close** ([fetcher_test.go:491-507](fetcher_test.go:491))
   - 특정 호스트의 모든 연결 닫기
   - Pool에서 제거 확인

#### Go 언어 개념 학습
- **동시성 제어**:
  - `sync.Mutex`: 공유 자원 보호
  - `Lock()` / `Unlock()`: Critical Section
  - `defer Unlock()`: 패닉 안전성
  - 경쟁 조건(Race Condition) 방지
- **네트워크 프로그래밍**:
  - `net.Conn` 인터페이스 재사용
  - 연결 풀링 패턴
  - 리소스 관리 (메모리 누수 방지)
- **에러 처리**:
  - `%w` 포맷: 에러 래핑 (error wrapping)
  - `errors.Is()` / `errors.As()`: 에러 체인 검사
  - Named return values: 명확한 반환값 문서화

#### 문제 발견: Transfer-Encoding: chunked
- **증상**: `example.org` 접속 시 프로그램 멈춤
- **원인 분석**:
  - 서버가 `Transfer-Encoding: chunked` 응답
  - `Content-Length` 헤더 없음
  - 현재 코드: `io.ReadAll()` 시도
  - Keep-alive라서 EOF 안 옴 → 무한 대기
- **HTTP Body 전송 방식**:
  1. `Content-Length`: 정확한 바이트 수 (✅ 구현 완료)
  2. `Transfer-Encoding: chunked`: 조각 전송 (⬜ 다음 작업)
  3. `Connection: close`: 연결 끊어서 끝 표시 (구식)

#### 다음 작업
- **Transfer-Encoding: chunked 구현**:
  - Chunked encoding 형식 이해
  - Chunk size 파싱
  - 여러 chunk 읽기
  - 마지막 chunk (0\r\n) 감지
  - Keep-Alive 유지하면서 body 읽기

#### 설계 결정
- **리팩토링 우선**: 코드 정리 후 새 기능 추가
- **Before/After 방식**: 학생이 직접 타이핑하며 학습
- **테스트 주도**: 기능 추가 전 테스트 작성

#### 최종 테스트 현황
- **총 44개 테스트** (43 pass, 1 skip)
  - parseHTML: 5개
  - NewURL: 8개
  - parsePort: 6개
  - parseHostPath: 6개
  - FileFetcher: 4개
  - DataFetcher: 6개
  - HTTPFetcher: 5개
  - ConnectionPool: 4개 (신규)
