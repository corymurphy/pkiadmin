# Issue Certificate

## needed

- ADCS Template
- KeyUsage

## steps

```csharp
KeyUsage keyUsage = dataTransformation.ParseKeyUsage(model.KeyUsage);
AdcsTemplate template = templateLogic.DiscoverTemplate
certificateProvider.CreateCsrKeyPair
ca.Sign
```


## certificateProvider.CreateCsrKeyPair

```csharp
CX509PrivateKey privateKey = CreatePrivateKey(cipher, keysize, api);
CreateCsrFromPrivateKey(subject, cipher, keysize, privateKey);
```

## 1. CreatePrivateKey

```csharp
CX509PrivateKey privateKey = CreatePrivateKey(cipher, keysize);
CX509CertificateRequestCertificate pkcs10 = NewCertificateRequestCrc(subject, privateKey);
pkcs10.Issuer = pkcs10.Subject;
pkcs10.NotBefore = DateTime.Now.AddDays(-1);
pkcs10.NotAfter = DateTime.Now.AddYears(20);
var sigoid = new CObjectId();
var alg = new Oid("SHA256");
sigoid.InitializeFromValue(alg.Value);
pkcs10.SignatureInformation.HashAlgorithm = sigoid;
pkcs10.Encode();

CX509Enrollment enrollment = new CX509Enrollment();

enrollment.InitializeFromRequest(pkcs10);

string csr = enrollment.CreateRequest(EncodingType.XCN_CRYPT_STRING_BASE64);
InstallResponseRestrictionFlags restrictionFlags = InstallResponseRestrictionFlags.AllowUntrustedCertificate;
enrollment.InstallResponse(restrictionFlags, csr, EncodingType.XCN_CRYPT_STRING_BASE64, string.Empty);

string pwd = secret.NewSecret(16);
string pfx = enrollment.CreatePFX(pwd, PFXExportOptions.PFXExportChainWithRoot, EncodingType.XCN_CRYPT_STRING_BASE64);
return new X509Certificate2(Convert.FromBase64String(pfx), pwd);
```

https://github.com/go-ole/go-ole
https://github.com/golang/go/wiki/WindowsDLLs
https://anubissec.github.io/How-To-Call-Windows-APIs-In-Golang/#
https://justen.codes/breaking-all-the-rules-using-go-to-call-windows-api-2cbfd8c79724?gi=1337f3df6dc9
https://www.thesubtlety.com/post/getting-started-golang-windows-apis/



## generating a referent id

```python
class PCLSID_ARRAY(NDRPOINTER):
    referent = (
        ('Data', CLSID_ARRAY),
    )

return data + NDRSTRUCT.getData(self, soFar)

# referent id

# This field is a serialization mark and it may have any value. For instance in our rpc implementation we are using running numbers. This field means a reference and the referenced value comes later according to the serialization rules. The value is meaningless.

# struct s1 { int f1; struct s2 * f2; int f3; } The above data is serialized as 1) f1 value 2) ref id for f2 3) f3 value 4) s2 contents. Hope this helps
```