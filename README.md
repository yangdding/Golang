# 파일 공유 서비스

Go로 구축된 안전한 웹 기반 파일 업로드 및 공유 서비스입니다.

## 주요 기능

- 드래그 앤 드롭 지원 웹 UI를 통한 파일 업로드
- 자동 공유 링크 생성
- curl/CLI를 위한 완전한 API 지원
- 포괄적인 보안 검증
- 배포 준비 완료 (Render, Docker 등)

## 보안 기능

- 파일 크기 제한 (최대 50MB)
- MIME 타입 검증
- 확장자 검증 (.jpg, .png, .gif, .pdf, .txt, .zip)
- 경로 탐색 공격 방지
- 파일명 새니타이징
- 보안 헤더 (X-Content-Type-Options, X-Frame-Options 등)
- API 접근을 위한 CORS 지원

## 지원 파일 형식

- **이미지**: JPG, PNG, GIF
- **문서**: PDF, TXT
- **압축**: ZIP


## 배포

### Render

1. GitHub에 코드 푸시
2. Render에서 새 Web Service 생성
3. GitHub 저장소 연결
4. Render가 자동으로 `render.yaml`을 감지하고 배포



## 프로젝트 구조

```
GOlang/
├── main.go              # 메인 애플리케이션
├── go.mod               # Go 모듈 의존성
├── go.sum               # 의존성 체크섬
├── Dockerfile           # Docker 설정
├── render.yaml          # Render 배포 설정
├── README.md            # 이 파일
└── uploads/             # 업로드된 파일 저장 디렉토리
```

## 보안 검증 세부사항

### 경로 탐색 공격 방지
- 모든 파일명은 `filepath.Base()`를 사용하여 새니타이징됨
- 파일 ID는 32자 16진수 문자열로 검증됨
- 절대 경로를 확인하여 파일이 업로드 디렉토리 내에 있는지 확인

### 파일 검증
1. **크기 검사**: 파일당 최대 50MB
2. **확장자 검사**: 허용된 확장자만 (.jpg, .png, .gif, .pdf, .txt, .zip)
3. **MIME 타입 검사**: Content-Type 헤더 검증
4. **파일명 새니타이징**: 경로 구분자 및 제어 문자 제거

### HTTP 보안 헤더
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 1; mode=block`
- `Referrer-Policy: strict-origin-when-cross-origin`



## 라이선스
MIT License


